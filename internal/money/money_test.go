package money

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseString(t *testing.T) {
	type args struct {
		s string
	}
	tests := map[string]struct {
		args    args
		want    M
		wantErr bool
	}{
		"float1": {
			args: args{"50.32"},
			want: M{
				Units: 50,
				Nanos: 320_000_000,
			},
			wantErr: false,
		},
		"float2": {
			args: args{"33.2"},
			want: M{
				Units: 33,
				Nanos: 200_000_000,
			},
			wantErr: false,
		},
		"float3": {
			args: args{"155.07"},
			want: M{
				Units: 155,
				Nanos: 70_000_000,
			},
			wantErr: false,
		},
		"float4": {
			args: args{"155.7"},
			want: M{
				Units: 155,
				Nanos: 700_000_000,
			},
			wantErr: false,
		},

		"neg_float1": {
			args: args{"-155.07"},
			want: M{
				Units: -155,
				Nanos: -70_000_000,
			},
			wantErr: false,
		},
		"neg_float2": {
			args: args{"-55.55"},
			want: M{
				Units: -55,
				Nanos: -550_000_000,
			},
			wantErr: false,
		},
		"zero_float1": {
			args: args{"00.00"},
			want: M{
				Units: 0,
				Nanos: 0,
			},
			wantErr: false,
		},
		"zero_float2": {
			args: args{"0.0"},
			want: M{
				Units: 0,
				Nanos: 0,
			},
			wantErr: false,
		},
		"zero_float3": {
			args: args{"0.00"},
			want: M{
				Units: 0,
				Nanos: 0,
			},
			wantErr: false,
		},
		"int1": {
			args: args{"5"},
			want: M{
				Units: 5,
				Nanos: 0,
			},
			wantErr: false,
		},
		"int2": {
			args: args{"325"},
			want: M{
				Units: 325,
				Nanos: 0,
			},
			wantErr: false,
		},
		"int3": {
			args: args{"1667"},
			want: M{
				Units: 1667,
				Nanos: 0,
			},
			wantErr: false,
		},
		"neg_int1": {
			args: args{"-5"},
			want: M{
				Units: -5,
				Nanos: 0,
			},
			wantErr: false,
		},
		"neg_int2": {
			args: args{"-325"},
			want: M{
				Units: -325,
				Nanos: 0,
			},
			wantErr: false,
		},
		"neg_int3": {
			args: args{"-1667"},
			want: M{
				Units: -1667,
				Nanos: 0,
			},
			wantErr: false,
		},
		"zero_int": {
			args: args{"0"},
			want: M{
				Units: 0,
				Nanos: 0,
			},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseString(tt.args.s)
			assert.Equal(t, tt.wantErr, (err != nil))
			assert.Equal(t, tt.want, got)
		})
	}
}
