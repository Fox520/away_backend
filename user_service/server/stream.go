package server

import (
	"encoding/json"
	"time"

	auth "github.com/Fox520/away_backend/auth"
	pb "github.com/Fox520/away_backend/user_service/github.com/Fox520/away_backend/user_service/pb"
)

func (server *UserServiceServer) StreamUserInfo(req *pb.StreamUserInfoRequest, stream pb.UserService_StreamUserInfoServer) error {
	meta := stream.Context().Value(auth.ContextMetaDataKey).(map[string]string)

	userId := meta[auth.ContextUIDKey]
	// Send initial data
	res, err := server.GetUser(stream.Context(), &pb.GetUserRequest{
		Id: userId,
	})
	if err != nil {
		return err
	}
	stream.Send(&pb.StreamUserInfoResponse{User: res.GetUser()})

	// We use this to avoid an extra db call
	// `UPDATE` event does not return previous value, thus store the original for later comparisons
	currentSubscriptionStatus := res.GetUser().SubscriptionStatus
	// NB: subscription status description is not in `users`, so events from the table will not include it
	// Store it in-memory for use in responses
	currentSubscriptionStatusDescription := res.GetUser().SubscriptionStatusDescription

	for {
		select {
		case n := <-server.usersDbListener.Notify:

			output := streamResult{}
			if err := json.Unmarshal([]byte(n.Extra), &output); err != nil {
				logger.Println(err)
				return err
			}
			extractedUser := output.Data
			// Build the response user struct
			// We can't directly unmarshal due to timestamp conversion issues
			resultUser := pb.AwayUser{
				Id:                 extractedUser.Id,
				UserName:           extractedUser.UserName,
				Email:              extractedUser.Email,
				Bio:                extractedUser.Bio,
				DeviceToken:        extractedUser.DeviceToken,
				Verified:           extractedUser.Verified,
				SubscriptionStatus: extractedUser.SubscriptionStatus,
				// Retrieve status description from db if it changed
				SubscriptionStatusDescription: currentSubscriptionStatusDescription,
			}
			if resultUser.Id != userId {
				continue
			}
			// Now we can make that call and update in-memory version
			if resultUser.SubscriptionStatus != currentSubscriptionStatus {

				currentSubscriptionStatus = resultUser.SubscriptionStatus
				currentSubscriptionStatusDescription = resultUser.SubscriptionStatusDescription

				// todo: create getUser func with direct db access
				res, err := server.GetUser(stream.Context(), &pb.GetUserRequest{
					Id: userId,
				})
				if err != nil {
					return err
				}
				resultUser = *res.GetUser()
			}
			if err := stream.Send(&pb.StreamUserInfoResponse{
				User: &resultUser,
			}); err != nil {
				logger.Println(err)
				stream.Context().Done()
				return err
			}

		case <-time.After(90 * time.Second):
			// fmt.Println("Received no events for 90 seconds, checking connection")
			go func() {
				server.usersDbListener.Ping()
			}()
		}
	}
}

type streamResult struct {
	Table  string        `json:"table"`
	Action string        `json:"action"`
	Data   streamUserObj `json:"data"`
}

// `createdAt` field is excluded for json decode to work; relevant error below
// parsing time "\"2022-03-06T16:32:03.258088\"" as "\"2006-01-02T15:04:05Z07:00\"": cannot parse "\"" as "Z07:00"
type streamUserObj struct {
	Id                 string `json:"id"`
	UserName           string `json:"username"`
	Email              string `json:"email"`
	Bio                string `json:"bio"`
	DeviceToken        string `json:"device_token"`
	Verified           bool   `json:"verified"`
	SubscriptionStatus string `json:"s_status"`
	ProfilePictureUrl  string `json:"profile_picture_url"`
}
