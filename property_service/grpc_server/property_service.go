package grpc_server

// TODO: package name that makes sense; figure out where to place this
import (
	"context"
	"errors"
	"log"
	"os"

	auth "github.com/Fox520/away_backend/auth"
	db "github.com/Fox520/away_backend/property_service/db/sqlc"
	pb "github.com/Fox520/away_backend/property_service/pb"
	propertyRepo "github.com/Fox520/away_backend/property_service/repository/property"
	user_pb "github.com/Fox520/away_backend/user_service/pb"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var logger = log.New(os.Stderr, "property_service: ", log.LstdFlags|log.Lshortfile)

type PropertyServiceServer struct {
	pb.UnimplementedPropertyServiceServer
	UserClient user_pb.UserServiceClient
	repo       propertyRepo.PropertyRepository
}

func NewPropertyServiceServer(userClient user_pb.UserServiceClient) *PropertyServiceServer {

	return &PropertyServiceServer{
		repo:       *propertyRepo.NewPropertyRepository(),
		UserClient: userClient,
	}

}

func (server *PropertyServiceServer) CreateProperty(ctx context.Context, request *pb.CreatePropertyRequest) (*pb.Property, error) {
	property := request.Property
	if property.Currency == "" || property.Title == "" || property.Photos == nil {
		return nil, errors.New("field requirements not met")
	}
	meta := ctx.Value(auth.ContextMetaDataKey).(map[string]string)
	userId := meta[auth.ContextUIDKey]

	property.UserID = userId

	arg := db.CreatePropertyParams{
		UserID:              property.UserID,
		PropertyTypeID:      int16(property.PropertyTypeID),
		PropertyCategoryID:  int16(property.PropertyCategoryID),
		PropertyUsageID:     int16(property.PropertyUsageID),
		Bedrooms:            int16(property.Bedrooms),
		Bathrooms:           int16(property.Bathrooms),
		Title:               property.Title,
		Currency:            property.Currency,
		Available:           property.Available,
		Price:               property.Price,
		Deposit:             property.Deposit,
		PetsAllowed:         property.PetsAllowed,
		FreeWifi:            property.FreeWifi,
		WaterIncluded:       property.WaterIncluded,
		ElectricityIncluded: property.ElectricityIncluded,
		Latitude:            property.Latitude,
		Longitude:           property.Longitude,
	}
	if property.Surburb != "" {
		arg.Surburb = &property.Surburb
	}
	if property.Town != "" {
		arg.Town = &property.Town
	}
	if property.Description != "" {
		arg.PDescription = &property.Description
	}
	if property.SharingPrice != 0 {
		arg.SharingPrice = &property.SharingPrice
	}
	result, err := server.repo.CreateProperty(ctx, arg)
	if err != nil {
		return nil, err
	}
	// Now add the photos
	for _, photo := range property.Photos {
		server.repo.CreatePropertyPhoto(ctx, db.CreatePropertyPhotoParams{
			PropertyID: result.ID,
			PUrl:       photo.Url,
		})
	}
	// Retrieve property (extra db call)
	return server.repo.GetProperty(ctx, result.ID)

}

func (server *PropertyServiceServer) UpdateProperty(context.Context, *pb.Property) (*pb.Property, error) {
	return nil, nil
}

func (server *PropertyServiceServer) DeleteProperty(ctx context.Context, dr *pb.DeletePropertyRequest) (*pb.DeletePropertyResponse, error) {
	meta := ctx.Value(auth.ContextMetaDataKey).(map[string]string)

	userId := meta[auth.ContextUIDKey]
	// Make sure it exists and belongs to requesting owner
	propId, err := uuid.Parse(dr.PropertyID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid arguments")
	}
	property, err := server.repo.GetMinimalProperty(ctx, propId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Property not found")
	}
	if userId != property.UserID {
		return nil, status.New(codes.PermissionDenied, "unauthorised request").Err()
	}
	err = server.repo.DeleteProperty(ctx, propId)
	if err != nil {
		return nil, err
	}
	return &pb.DeletePropertyResponse{
		Status: true,
	}, nil
}
