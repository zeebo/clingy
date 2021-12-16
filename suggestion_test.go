package clingy

import "testing"

func Test_levenshteinDistance(t *testing.T) {
	type args struct {
		a string
		b string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "different strings 1",
			args: args{
				a: "Sitting",
				b: "Kitten",
			},
			want: 3,
		},
		{
			name: "different strings 2",
			args: args{
				a: "Sunday",
				b: "Saturday",
			},
			want: 3,
		},
		{
			name: "strings are the same",
			args: args{
				a: "Sunday",
				b: "Sunday",
			},
			want: 0,
		},
		{
			name: "different strings 3",
			args: args{
				a: "abcdefgh",
				b: "abcxdefghi",
			},
			want: 2,
		},
		{
			name: "with numbers",
			args: args{
				a: "123456",
				b: "234156",
			},
			want: 2,
		},
		{
			name: "with numbers 2",
			args: args{
				a: "123657",
				b: "123",
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := levenshteinDistance(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("EditDistance() = %v, want %v", got, tt.want)
			}
		})
	}
}
