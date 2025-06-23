package security

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLuhnPredicat(t *testing.T) {
	type args struct {
		numeralID string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test 79927398713 -> True",
			args: args{numeralID: "79927398713"},
			want: true,
		},
		{
			name: "test 79927398712 -> False",
			args: args{numeralID: "79927398712"},
			want: false,
		},
		{
			name: "test 1234567897 -> True",
			args: args{numeralID: "1234567897"},
			want: true,
		},
		{
			name: "test 1234567890 -> False",
			args: args{numeralID: "1234567890"},
			want: false,
		},
		{
			name: "test 4532015112830366 -> True",
			args: args{numeralID: "4532015112830366"},
			want: true,
		},
		{
			name: "test 4532015112830367 -> False",
			args: args{numeralID: "4532015112830367"},
			want: false,
		},
		{
			name: "test 6011111111111117 -> True",
			args: args{numeralID: "6011111111111117"},
			want: true,
		},
		{
			name: "test 6011111111111118 -> False",
			args: args{numeralID: "6011111111111118"},
			want: false,
		},
		{
			name: "test 123455 -> True",
			args: args{numeralID: "123455"},
			want: true,
		},
		{
			name: "test 40312332 -> False",
			args: args{numeralID: "40312332"},
			want: true,
		},
		{
			name: "test 9278923470 -> True",
			args: args{numeralID: "9278923470"},
			want: true,
		},
		{
			name: "test 12345678903 -> True",
			args: args{numeralID: "12345678903"},
			want: true,
		},
		{
			name: "test 346436439 -> True",
			args: args{numeralID: "346436439"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, LuhnAlgoPredicat(tt.args.numeralID), "LuhnAlgoPredicat(%v)", tt.args.numeralID)
		})
	}
}
