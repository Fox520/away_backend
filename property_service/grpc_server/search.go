package grpc_server

import (
	"context"
	"sync"

	"github.com/kr/pretty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"googlemaps.github.io/maps"
)

var mapsClient *maps.Client
var once sync.Once

// Retrieves locations matched with input
func performSearch(ctx context.Context, sessionToken maps.PlaceAutocompleteSessionToken, input *string, country string) ([]maps.AutocompletePrediction, error) {
	if getMapsClient() == nil {
		return nil, status.Error(codes.Internal, "Cannot perform search now.")
	}
	r := &maps.PlaceAutocompleteRequest{
		Input: *input,
		// Offset:       0,
		StrictBounds: false,
		SessionToken: sessionToken,
		Components:   make(map[maps.Component][]string),
	}
	r.Components[maps.ComponentCountry] = append(r.Components[maps.ComponentCountry], country)
	resp, err := getMapsClient().PlaceAutocomplete(ctx, r)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pretty.Println(resp)
	return resp.Predictions, nil

}
func getLocation(ctx context.Context, sessionToken maps.PlaceAutocompleteSessionToken, placeId *string) (maps.LatLng, error) {
	if getMapsClient() == nil {
		return maps.LatLng{}, status.Error(codes.Internal, "Cannot perform search now.")
	}
	r := &maps.PlaceDetailsRequest{
		PlaceID:      *placeId,
		SessionToken: sessionToken,
	}
	resp, err := getMapsClient().PlaceDetails(ctx, r)
	if err != nil {
		return maps.LatLng{}, status.Error(codes.Internal, err.Error())
	}

	pretty.Println(resp)

	return resp.Geometry.Location, nil

}

func getMapsClient() *maps.Client {
	once.Do(func() {
		client, err := maps.NewClient(maps.WithAPIKey("AIzaSyDNfrYn4wsYNpoxENEEuscxOAkcDUxg3lA")) //maps.NewClient(maps.WithAPIKey(os.Getenv("MAPS_KEY")))
		if err != nil {
			logger.Println(err)
			mapsClient = nil
		} else {
			mapsClient = client

		}
	})
	return mapsClient
}
