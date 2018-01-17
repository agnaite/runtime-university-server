package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/heroku/runtime-university-server/spec"
	"golang.org/x/net/context"
)

var jsonDBFile = "testdata/route_guide_db.json"

type RouteGuideServer struct {
	savedFeatures []*spec.Feature
	routeNotes    map[string][]*spec.RouteNote
}

func NewRouteGuideServer() *RouteGuideServer {
	s := new(RouteGuideServer)
	s.loadFeatures(jsonDBFile)
	s.routeNotes = make(map[string][]*spec.RouteNote)
	return s
}

// GetFeature returns the feature at the given point.
func (s *RouteGuideServer) GetFeature(ctx context.Context, point *spec.Point) (*spec.Feature, error) {
	for _, feature := range s.savedFeatures {
		if proto.Equal(feature.Location, point) {
			return feature, nil
		}
	}
	// No feature was found, return an unnamed feature
	return &spec.Feature{Location: point}, nil
}

// ListFeatures lists all features contained within the given bounding Rectangle.
func (s *RouteGuideServer) ListFeatures(rect *spec.Rectangle, stream spec.RouteGuide_ListFeaturesServer) error {
	for _, feature := range s.savedFeatures {
		if inRange(feature.Location, rect) {
			if err := stream.Send(feature); err != nil {
				return err
			}
		}
	}
	return nil
}

// RecordRoute records a route composited of a sequence of points.
//
// It gets a stream of points, and responds with statistics about the "trip":
// number of points,  number of known features visited, total distance traveled, and
// total time spent.
func (s *RouteGuideServer) RecordRoute(stream spec.RouteGuide_RecordRouteServer) error {
	var pointCount, featureCount, distance int32
	var lastPoint *spec.Point
	startTime := time.Now()
	for {
		point, err := stream.Recv()
		if err == io.EOF {
			endTime := time.Now()
			return stream.SendAndClose(&spec.RouteSummary{
				PointCount:   pointCount,
				FeatureCount: featureCount,
				Distance:     distance,
				ElapsedTime:  int32(endTime.Sub(startTime).Seconds()),
			})
		}
		if err != nil {
			return err
		}
		pointCount++
		for _, feature := range s.savedFeatures {
			if proto.Equal(feature.Location, point) {
				featureCount++
			}
		}
		if lastPoint != nil {
			distance += calcDistance(lastPoint, point)
		}
		lastPoint = point
	}
}

// RouteChat receives a stream of message/location pairs, and responds with a stream of all
// previous messages at each of those locations.
func (s *RouteGuideServer) RouteChat(stream spec.RouteGuide_RouteChatServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		key := serialize(in.Location)
		if _, present := s.routeNotes[key]; !present {
			s.routeNotes[key] = []*spec.RouteNote{in}
		} else {
			s.routeNotes[key] = append(s.routeNotes[key], in)
		}
		for _, note := range s.routeNotes[key] {
			if err := stream.Send(note); err != nil {
				return err
			}
		}
	}
}

// LoadFeatures loads features from a JSON file.
func (s *RouteGuideServer) loadFeatures(filePath string) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to load default features: %v", err)
	}
	if err := json.Unmarshal(file, &s.savedFeatures); err != nil {
		log.Fatalf("Failed to load default features: %v", err)
	}
}

func toRadians(num float64) float64 {
	return num * math.Pi / float64(180)
}

// calcDistance calculates the distance between two points using the "haversine" formula.
// This code was taken from http://www.movable-type.co.uk/scripts/latlong.html.
func calcDistance(p1 *spec.Point, p2 *spec.Point) int32 {
	const CordFactor float64 = 1e7
	const R float64 = float64(6371000) // metres
	lat1 := float64(p1.Latitude) / CordFactor
	lat2 := float64(p2.Latitude) / CordFactor
	lng1 := float64(p1.Longitude) / CordFactor
	lng2 := float64(p2.Longitude) / CordFactor
	φ1 := toRadians(lat1)
	φ2 := toRadians(lat2)
	Δφ := toRadians(lat2 - lat1)
	Δλ := toRadians(lng2 - lng1)

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*
			math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := R * c
	return int32(distance)
}

func inRange(point *spec.Point, rect *spec.Rectangle) bool {
	left := math.Min(float64(rect.Lo.Longitude), float64(rect.Hi.Longitude))
	right := math.Max(float64(rect.Lo.Longitude), float64(rect.Hi.Longitude))
	top := math.Max(float64(rect.Lo.Latitude), float64(rect.Hi.Latitude))
	bottom := math.Min(float64(rect.Lo.Latitude), float64(rect.Hi.Latitude))

	if float64(point.Longitude) >= left &&
		float64(point.Longitude) <= right &&
		float64(point.Latitude) >= bottom &&
		float64(point.Latitude) <= top {
		return true
	}
	return false
}

func serialize(point *spec.Point) string {
	return fmt.Sprintf("%d %d", point.Latitude, point.Longitude)
}
