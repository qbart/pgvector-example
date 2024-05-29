package web

import (
	"SoftKiwiGames/go-web-template/accounts/dto"
	"context"
)

type MockAccount struct {
}

func (m *MockAccount) GetProfile(ctx context.Context) (dto.User, error) {
	return dto.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		AvatarURL: "https://www.gravatar.com/avatar/e5c7d8f1e0a9d046b1302c1e45624f45",
	}, nil
}
