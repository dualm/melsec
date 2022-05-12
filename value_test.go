package melsec

import "testing"

func Test_splitComponetName(t *testing.T) {
	type args struct {
		component string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{
			name: "1",
			args: args{
				component: "X0",
			},
			want:  "X",
			want1: "0",
		},
		{
			name: "2",
			args: args{
				component: "SB2",
			},
			want:  "SB",
			want1: "2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := splitComponetName(tt.args.component)
			if got != tt.want {
				t.Errorf("splitComponetName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("splitComponetName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
