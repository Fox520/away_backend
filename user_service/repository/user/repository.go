package user

import (
	"context"
	"log"

	"github.com/Fox520/away_backend/auth"
	db_collections "github.com/Fox520/away_backend/user_service/db"
	db "github.com/Fox520/away_backend/user_service/db/sqlc"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserRepository struct {
	query *db.Queries
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		query: db.New(db_collections.GetAwayDB()),
	}
}

func (repo *UserRepository) CreateUser(ctx context.Context, arg db.CreateUserParams) (*db.User, error) {
	user, err := repo.query.CreateUser(ctx, arg)

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
			switch err.Code {
			case "23505": // unique_violation
				log.Println("User create duplicate:", err.Code.Name())
				return nil, status.Error(codes.AlreadyExists, "User already exists")
			}
		}
		log.Println("user insert error: ", err)
		return nil, status.Error(codes.Internal, "Could not create user")
	}
	return &user, nil
}

func (repo *UserRepository) GetFullUser(ctx context.Context, userId string) (*db.GetFullUserRow, error) {
	user, err := repo.query.GetFullUser(ctx, userId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "User not found")
	}
	return &user, nil
}

func (repo *UserRepository) GetMinimalUser(ctx context.Context, userId string) (*db.GetMinimalUserRow, error) {
	user, err := repo.query.GetMinimalUser(ctx, userId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "User not found")
	}
	return &user, nil
}

func (repo *UserRepository) SetUsername(ctx context.Context, arg db.SetUsernameParams) (db.User, error) {
	return repo.query.SetUsername(ctx, arg)
}

func (repo *UserRepository) SetDeviceToken(ctx context.Context, arg db.SetDeviceTokenParams) (db.User, error) {
	return repo.query.SetDeviceToken(ctx, arg)
}

func (repo *UserRepository) SetBio(ctx context.Context, arg db.SetBioParams) (db.User, error) {
	return repo.query.SetBio(ctx, arg)
}

func (repo *UserRepository) SetProfilePictureUrl(ctx context.Context, arg db.SetProfilePictureUrlParams) (db.User, error) {
	return repo.query.SetProfilePictureUrl(ctx, arg)
}

func (repo *UserRepository) DeleteUser(ctx context.Context, userId string) error {
	err := repo.query.DeleteUser(ctx, userId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
			switch err.Code {
			case "02000": // no_data
				return status.Error(codes.NotFound, "User not found")
			}
		}
		log.Println("user delete error: ", err)
		return status.Error(codes.Internal, "Could not delete user")
	}
	err = auth.GetFirebaseAuthClient().DeleteUser(context.Background(), userId)
	return err

}
