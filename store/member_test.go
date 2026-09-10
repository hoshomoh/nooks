package store

import (
	"errors"
	"testing"
	"time"
)

var createdAt = time.Date(2026, time.March, 3, 9, 30, 0, 0, time.UTC)

// anna is a Member good enough for most cases.
func anna() CreateMemberParams {
	return CreateMemberParams{
		UID:          "mem_anna",
		Name:         "Anna",
		Email:        "anna@brunnen.lan",
		Role:         RoleAdmin,
		PasswordHash: "$2a$10$notarealhashbutlongenoughtostore",
		CreatedAt:    createdAt,
	}
}

func TestCreateAndReadMember(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)

			created, err := s.CreateMember(t.Context(), anna())
			if err != nil {
				t.Fatalf("CreateMember: %v", err)
			}
			if created.ID == 0 {
				t.Error("ID = 0, want the database to have assigned one")
			}
			if !created.IsAdmin() {
				t.Error("IsAdmin() = false, want true for the first Member")
			}
			if !created.CreatedAt.Equal(createdAt) {
				t.Errorf("CreatedAt = %v, want %v", created.CreatedAt, createdAt)
			}
			if !created.LastSignedInAt.IsZero() {
				t.Error("LastSignedInAt is set on a Member who has never signed in")
			}

			found, err := s.MemberByEmail(t.Context(), "anna@brunnen.lan")
			if err != nil {
				t.Fatalf("MemberByEmail: %v", err)
			}
			if found.ID != created.ID {
				t.Errorf("MemberByEmail gave ID %d, want %d", found.ID, created.ID)
			}

			byUID, err := s.MemberByUID(t.Context(), "mem_anna")
			if err != nil {
				t.Fatalf("MemberByUID: %v", err)
			}
			if byUID.ID != created.ID {
				t.Errorf("MemberByUID gave ID %d, want %d", byUID.ID, created.ID)
			}
		})
	}
}

// Nobody thinks of their address as case-sensitive, so sign-in must not either.
func TestMemberEmailIsCaseInsensitive(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)

			params := anna()
			params.Email = "  Anna@Brunnen.LAN  "
			if _, err := s.CreateMember(t.Context(), params); err != nil {
				t.Fatalf("CreateMember: %v", err)
			}

			found, err := s.MemberByEmail(t.Context(), "ANNA@brunnen.lan")
			if err != nil {
				t.Fatalf("MemberByEmail with different case: %v", err)
			}
			if found.Email != "anna@brunnen.lan" {
				t.Errorf("Email stored as %q, want it normalised", found.Email)
			}
		})
	}
}

func TestCreateMemberRejectsADuplicateEmail(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			if _, err := s.CreateMember(t.Context(), anna()); err != nil {
				t.Fatalf("CreateMember: %v", err)
			}

			second := anna()
			second.UID = "mem_other"
			second.Name = "Someone else"
			if _, err := s.CreateMember(t.Context(), second); !errors.Is(err, ErrEmailTaken) {
				t.Errorf("CreateMember with a taken email = %v, want ErrEmailTaken", err)
			}
		})
	}
}

func TestMemberNotFound(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			if _, err := s.MemberByEmail(t.Context(), "nobody@brunnen.lan"); !errors.Is(err, ErrNotFound) {
				t.Errorf("MemberByEmail for a stranger = %v, want ErrNotFound", err)
			}
			if _, err := s.MemberByUID(t.Context(), "mem_nobody"); !errors.Is(err, ErrNotFound) {
				t.Errorf("MemberByUID for a stranger = %v, want ErrNotFound", err)
			}
		})
	}
}

// First run is complete once there is a Member, so the count is what the app asks.
func TestCountMembers(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)

			count, err := s.CountMembers(t.Context())
			if err != nil {
				t.Fatalf("CountMembers: %v", err)
			}
			if count != 0 {
				t.Errorf("CountMembers on a fresh Instance = %d, want 0", count)
			}

			if _, err := s.CreateMember(t.Context(), anna()); err != nil {
				t.Fatalf("CreateMember: %v", err)
			}
			if count, err = s.CountMembers(t.Context()); err != nil || count != 1 {
				t.Errorf("CountMembers = %d, %v; want 1, nil", count, err)
			}
		})
	}
}

// Replacing a temporary password is exactly what clears the must-change flag.
func TestSetMemberPasswordClearsMustChange(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)

			params := anna()
			params.MustChangePassword = true
			created, err := s.CreateMember(t.Context(), params)
			if err != nil {
				t.Fatalf("CreateMember: %v", err)
			}
			if !created.MustChangePassword {
				t.Fatal("MustChangePassword = false, want true for a temporary password")
			}

			if err := s.SetMemberPassword(t.Context(), created.ID, "$2a$10$areplacementhashvaluehere"); err != nil {
				t.Fatalf("SetMemberPassword: %v", err)
			}

			after, err := s.MemberByID(t.Context(), created.ID)
			if err != nil {
				t.Fatalf("MemberByID: %v", err)
			}
			if after.MustChangePassword {
				t.Error("MustChangePassword is still true after the password was replaced")
			}
			if after.PasswordHash == created.PasswordHash {
				t.Error("PasswordHash is unchanged")
			}
		})
	}
}

func TestSetMemberPasswordForAStranger(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			err := d.open(t).SetMemberPassword(t.Context(), 4242, "$2a$10$whatever")
			if !errors.Is(err, ErrNotFound) {
				t.Errorf("SetMemberPassword for a stranger = %v, want ErrNotFound", err)
			}
		})
	}
}

func TestMarkMemberSignedIn(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			created, err := s.CreateMember(t.Context(), anna())
			if err != nil {
				t.Fatalf("CreateMember: %v", err)
			}

			signedIn := createdAt.Add(48 * time.Hour)
			if err := s.MarkMemberSignedIn(t.Context(), created.ID, signedIn); err != nil {
				t.Fatalf("MarkMemberSignedIn: %v", err)
			}

			after, err := s.MemberByID(t.Context(), created.ID)
			if err != nil {
				t.Fatalf("MemberByID: %v", err)
			}
			if !after.LastSignedInAt.Equal(signedIn) {
				t.Errorf("LastSignedInAt = %v, want %v", after.LastSignedInAt, signedIn)
			}
		})
	}
}
