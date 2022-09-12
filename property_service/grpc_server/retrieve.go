package grpc_server

import (
	"context"
	"io"

	auth "github.com/Fox520/away_backend/auth"
	db "github.com/Fox520/away_backend/property_service/db/sqlc"
	pb "github.com/Fox520/away_backend/property_service/pb"
	user_pb "github.com/Fox520/away_backend/user_service/pb"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"googlemaps.github.io/maps"
)

// Fetch property from db
func (server *PropertyServiceServer) GetSingleProperty(ctx context.Context, pr *pb.GetSinglePropertyRequest) (*pb.SinglePropertyResponse, error) {

	var propertyResponse pb.SinglePropertyResponse

	propertyId, err := uuid.Parse(pr.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid property id")
	}

	result, err := server.repo.GetProperty(ctx, propertyId) // fetchAndCacheSingleProperty(ctx, pr.Id)
	if err != nil {
		return nil, err
	}
	// Fixes: invalid memory address or nil pointer dereference
	// propertyResponse.Property = &pb.Property{}
	propertyResponse.Property = result

	// Attempt adding user profile
	meta := ctx.Value(auth.ContextMetaDataKey).(map[string]string)

	token := meta[auth.ContextTokenKey]
	md := metadata.New(map[string]string{"token": token})
	requestContext := metadata.NewOutgoingContext(ctx, md)
	res, err := server.UserClient.GetUser(requestContext, &user_pb.GetUserRequest{Id: propertyResponse.Property.UserID})
	if err == nil {
		propertyResponse.Owner = res
	} else {
		logger.Println(err)
	}
	return &propertyResponse, nil
}

func (server *PropertyServiceServer) GetMultipleProperties(stream pb.PropertyService_GetMultiplePropertiesServer) error {
	meta := stream.Context().Value(auth.ContextMetaDataKey).(map[string]string)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			logger.Println(err)
			return err
		}
		radius := req.Radius
		if radius < 1 {
			radius = 5
		}
		// Convert radius (km) to meters
		radius = radius * 1000
		properties, err := server.repo.GetPropertiesWithinRadius(stream.Context(), db.GetPropertiesWithinRadiusParams{
			LlToEarth:   float64(req.Latitude),
			LlToEarth_2: float64(req.Longitude),
			EarthBox:    float64(radius),
		})
		if err != nil {
			return err
		}

		var owners map[string]*user_pb.GetUserResponse

		var responseProperties []*pb.SinglePropertyResponse

		for _, property := range properties {
			var response pb.SinglePropertyResponse
			// Avoid retrieving already fetched profiles
			if owners[property.UserID] == nil {
				// Attempt adding user profile
				token := meta[auth.ContextTokenKey]
				md := metadata.New(map[string]string{"token": token})
				requestContext := metadata.NewOutgoingContext(stream.Context(), md)
				res, err := server.UserClient.GetUser(requestContext, &user_pb.GetUserRequest{Id: property.UserID})
				if err == nil {
					response.Owner = res
				} else {
					logger.Println(err)
				}
			} else {
				response.Owner = owners[property.UserID]
			}
			response.Property = property

			responseProperties = append(responseProperties, &response)

		}
		if err := stream.Send(&pb.GetMultiplePropertyResponse{Response: responseProperties}); err != nil {
			return err
		}
	}
}

func (server *PropertyServiceServer) GetMinimalInfoProperties(stream pb.PropertyService_GetMinimalInfoPropertiesServer) error {
	meta := stream.Context().Value(auth.ContextMetaDataKey).(map[string]string)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			logger.Println(err)
			return err
		}
		radius := req.Radius
		if radius < 1 {
			radius = 5
		}
		// Convert radius (km) to meters
		radius = radius * 1000
		properties, err := server.repo.GetMinimalInfoPropertiesWithinRadius(stream.Context(), db.GetMinimalInfoPropertiesWithinRadiusParams{
			LlToEarth:   float64(req.Latitude),
			LlToEarth_2: float64(req.Longitude),
			EarthBox:    float64(radius),
		}, false)
		if err != nil {
			return err
		}

		var owners map[string]*user_pb.GetUserResponse
		var responseProperties []*pb.SingleMinimalProperty

		for _, property := range properties {
			var singleMinimalProperty pb.SingleMinimalProperty
			// Avoid retrieving already fetched profiles
			if owners[property.UserID] == nil {
				// Attempt adding user profile
				token := meta[auth.ContextTokenKey]
				md := metadata.New(map[string]string{"token": token})
				requestContext := metadata.NewOutgoingContext(stream.Context(), md)
				res, err := server.UserClient.GetUser(requestContext, &user_pb.GetUserRequest{Id: property.UserID})
				if err == nil {
					singleMinimalProperty.Owner = res
				} else {
					logger.Println(err)
				}
			} else {
				singleMinimalProperty.Owner = owners[property.UserID]
			}
			singleMinimalProperty.Property = property

			responseProperties = append(responseProperties, &singleMinimalProperty)
		}
		if err := stream.Send(&pb.GetMinimalPropertiesResponse{SingleMinimalProperties: responseProperties}); err != nil {
			return err
		}
	}

}

/*
Just like GetMinimalInfoProperties but for promoted properties
*/
func (server *PropertyServiceServer) GetPromotedProperties(stream pb.PropertyService_GetPromotedPropertiesServer) error {
	meta := stream.Context().Value(auth.ContextMetaDataKey).(map[string]string)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			logger.Println(err)
			return err
		}
		radius := req.Radius
		if radius < 1 {
			radius = 5
		}
		// Convert radius (km) to meters
		radius = radius * 1000
		properties, err := server.repo.GetMinimalInfoPropertiesWithinRadius(stream.Context(), db.GetMinimalInfoPropertiesWithinRadiusParams{
			LlToEarth:   float64(req.Latitude),
			LlToEarth_2: float64(req.Longitude),
			EarthBox:    float64(radius),
		}, true)
		if err != nil {
			return err
		}

		var owners map[string]*user_pb.GetUserResponse
		var responseProperties []*pb.SingleMinimalProperty

		for _, property := range properties {
			var singleMinimalProperty pb.SingleMinimalProperty
			// Avoid retrieving already fetched profiles
			if owners[property.UserID] == nil {
				// Attempt adding user profile
				token := meta[auth.ContextTokenKey]
				md := metadata.New(map[string]string{"token": token})
				requestContext := metadata.NewOutgoingContext(stream.Context(), md)
				res, err := server.UserClient.GetUser(requestContext, &user_pb.GetUserRequest{Id: property.UserID})
				if err == nil {
					singleMinimalProperty.Owner = res
				} else {
					logger.Println(err)
				}
			} else {
				singleMinimalProperty.Owner = owners[property.UserID]
			}
			singleMinimalProperty.Property = property

			responseProperties = append(responseProperties, &singleMinimalProperty)
		}

		if err := stream.Send(&pb.PromotedResponse{Properties: responseProperties}); err != nil {
			return err
		}
	}

}

func (server *PropertyServiceServer) GetFeaturedAreas(ctx context.Context, tr *pb.FeaturedAreasRequest) (*pb.FeaturedAreasResponse, error) {
	var response pb.FeaturedAreasResponse
	results, err := server.repo.GetFeaturedAreas(ctx, tr.Country)
	if err != nil {
		return nil, err
	}
	areas := make([]*pb.FeaturedArea, 0)
	for _, r := range results {
		areas = append(areas, &pb.FeaturedArea{Title: r.Title, PhotoURL: r.PhotoUrl, Latitude: r.Latitude, Longitude: r.Longitude})
	}
	response.FeaturedAreas = areas
	return &response, nil
}

func (server *PropertyServiceServer) LocationSearch(stream pb.PropertyService_LocationSearchServer) error {
	sessionToken := maps.NewPlaceAutocompleteSessionToken()
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			logger.Println(err)
			return err
		}
		switch req := req.SearchOneof.(type) {
		case *pb.LocationSearchRequest_Query:
			predictions, err := performSearch(stream.Context(), sessionToken, &req.Query.Text, req.Query.CountryCode)
			if err != nil {
				return err
			}
			// Map each prediction to a LocationAutocomplete
			autocompletes := make([]*pb.LocationAutocomplete, len(predictions))
			for index, prediction := range predictions {
				autocompletes[index] = &pb.LocationAutocomplete{
					Title:         prediction.StructuredFormatting.MainText,
					SecondaryText: prediction.StructuredFormatting.SecondaryText,
					PlaceID:       prediction.PlaceID,
				}

			}
			// Send off the autocomplete slice
			response := pb.LocationSearchResponse{ResponseOneof: &pb.LocationSearchResponse_AutocompleteResponse{AutocompleteResponse: &pb.LocationAutocompleteResponse{
				Responses: autocompletes,
			}}}
			if err := stream.Send(&response); err != nil {
				return err
			}

		case *pb.LocationSearchRequest_Details:
			location, err := getLocation(stream.Context(), sessionToken, &req.Details.PlaceID)
			if err != nil {
				return err
			}
			stream.Send(&pb.LocationSearchResponse{
				ResponseOneof: &pb.LocationSearchResponse_Details{
					Details: &pb.LocationDetails{
						Latitude:  float32(location.Lat),
						Longitude: float32(location.Lng),
						PlaceID:   req.Details.PlaceID,
					},
				},
			})
			stream.Context().Done()
			return nil
		}

	}
}
