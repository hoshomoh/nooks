package store

import (
	"errors"
	"fmt"
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

/*
What removing a Member takes with them.

Not a wish, a record. `list.owner_id` and `item.added_by_id` are both ON DELETE CASCADE,
so removing somebody deletes every List they started and every Item they ever added,
including the ones on other people's Lists that those people are still using.

Pinned here because two things in this repository assume the opposite. The Settings
dialog told an Admin that what somebody added stays on its lists, and `rowNamesFor` has
a branch for a row whose Member is gone — which `added_by_id` can never reach, because
the row goes with them. Whichever way that is settled, it should be settled on purpose
and this test should fail when it is.
*/
func TestRemovingAMemberTakesTheirListsAndTheirItems(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			anna := newMember(t, s)
			jonas := addMember(t, s, "mem_jonas", "Jonas", "jonas@brunnen.lan")

			hers := makeList(t, s, anna, "list_hers", "Groceries", SharingInstance)
			his := makeList(t, s, jonas, "list_his", "Flat jobs", SharingInstance)
			addItem(t, s, hers, anna, "item_milk", "Milk")
			addItem(t, s, his, anna, "item_bins", "Bins")
			addItem(t, s, his, jonas, "item_recycling", "Recycling")

			if err := s.DeleteMember(t.Context(), anna.ID); err != nil {
				t.Fatalf("DeleteMember: %v", err)
			}

			// The List she owned, which the whole Instance could see.
			if _, err := s.ListByUID(t.Context(), hers.UID); !errors.Is(err, ErrNotFound) {
				t.Errorf("her List is %v, and this test is what says that is deliberate", err)
			}

			// His List survives. What she put on it does not.
			items, err := s.ItemsOnList(t.Context(), his.ID)
			if err != nil {
				t.Fatalf("ItemsOnList: %v", err)
			}
			left := make([]string, 0, len(items))
			for _, item := range items {
				left = append(left, item.Label)
			}
			if fmt.Sprint(left) != "[Recycling]" {
				t.Errorf("his List holds %v, want only what he added himself", left)
			}
		})
	}
}

/*
Taking somebody else's email by editing a profile is refused the same way as by creating
an account.

Both are the database's unique index reporting itself, and both are read back out of the
driver's error text — which SQLite and Postgres word differently, so the reading is
checked on each. Creating had a test and editing had none. Without it a Member who
retypes an address somebody already has is handed an internal failure with the
constraint name in it, rather than being told the address is taken.
*/
func TestSetMemberProfileRejectsATakenEmail(t *testing.T) {
	for _, d := range drivers() {
		t.Run(d.name, func(t *testing.T) {
			s := d.open(t)
			if _, err := s.CreateMember(t.Context(), anna()); err != nil {
				t.Fatalf("CreateMember: %v", err)
			}

			other := anna()
			other.UID = "mem_jonas"
			other.Name = "Jonas"
			other.Email = "jonas@brunnen.lan"
			jonas, err := s.CreateMember(t.Context(), other)
			if err != nil {
				t.Fatalf("CreateMember: %v", err)
			}

			err = s.SetMemberProfile(t.Context(), jonas.ID, "Jonas", anna().Email)
			if !errors.Is(err, ErrEmailTaken) {
				t.Errorf("SetMemberProfile onto a taken email = %v, want ErrEmailTaken", err)
			}

			// His own is still his: a refused write changes nothing.
			read, err := s.MemberByID(t.Context(), jonas.ID)
			if err != nil {
				t.Fatalf("MemberByID: %v", err)
			}
			if read.Email != "jonas@brunnen.lan" {
				t.Errorf("Email = %q, want the one he had", read.Email)
			}
		})
	}
}
