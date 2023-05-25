package security

import (
	"fmt"
	"testing"
)

func TestPubSubAuth(t *testing.T) {
	// create a new authentication instance with a single ACL that allows all subscriptions
	auth := NewAuthentication([]Acl{
		{"a/b", CanSubPub},
		{"a/c", CanSubPub},
		{"a/c/d", CanSub},
		{"a/c/b", CanPub},
		{"a/s/#", CanSubPub},
		{"a/+/h", CanSub},
		{"a/d/h", CanSubPub},
	})

	tests := []struct {
		input  string
		canSub bool
		canPub bool
	}{
		{"a/s", false, false},
		{"f/dd/s/h", false, false},
		{"a/s/2", true, true},
		{"a/c/h", true, false},
		{"a/s/+", true, true},
		{"a/d/h", true, true},
		{"a/m/h/111", false, false},
		{"a/c/b", false, true},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("input=%v", test.input), func(t *testing.T) {
			canPub := auth.CanPub(test.input)
			canSub := auth.CanSub(test.input)
			if canPub != test.canPub {
				t.Errorf("expected canPub %v, but got %v", test.canPub, canPub)
			}
			if canSub != test.canSub {
				t.Errorf("expected canSub %v, but got %v", test.canSub, canSub)
			}
		})
	}
}

func TestDefaultAuthenticationProvider(t *testing.T) {
	// create a new authentication provider with two users
	users := []User{{"alice", "password1"}, {"bob", "password2"}}
	authProvider := NewStaticUserListAuthProvider(users)

	// test a valid authentication
	auth := authProvider.Authenticate("alice", "password1")
	if auth == nil {
		t.Errorf("expected Authenticate to return non-nil authentication for valid user")
	}

	// test an invalid username
	auth = authProvider.Authenticate("eve", "password")
	if auth != nil {
		t.Errorf("expected Authenticate to return nil authentication for invalid username")
	}

	// test an invalid password
	auth = authProvider.Authenticate("bob", "password")
	if auth != nil {
		t.Errorf("expected Authenticate to return nil authentication for invalid password")
	}
}
