package user

import "testing"

func TestInviteValidate(t *testing.T) {
	tests := []struct {
		name string
		in   InviteInput
		ok   bool
	}{
		{"ок", InviteInput{Name: " Анна ", Email: " Anna@Example.com ", Role: "manager"}, true},
		{"пустое имя", InviteInput{Name: " ", Email: "a@example.com", Role: "manager"}, false},
		{"битый email", InviteInput{Name: "Анна", Email: "anna@", Role: "manager"}, false},
		{"email с именем", InviteInput{Name: "Анна", Email: "Анна <a@example.com>", Role: "manager"}, false},
		{"неизвестная роль", InviteInput{Name: "Анна", Email: "a@example.com", Role: "root"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := tt.in
			err := in.Validate()
			if (err == nil) != tt.ok {
				t.Fatalf("Validate() = %v, want ok=%v", err, tt.ok)
			}
			if tt.ok && (in.Name != "Анна" || in.Email != "anna@example.com") {
				t.Errorf("не нормализовано: %+v", in)
			}
		})
	}
}
