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
			got, got1 := splitComponentName(tt.args.component)
			if got != tt.want {
				t.Errorf("splitComponentName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("splitComponentName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_splitComponetName1(t *testing.T) {
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
				component: "XA0",
			},
			want:  "X",
			want1: "A0",
		},
		{
			name: "2",
			args: args{
				component: "XAA0",
			},
			want:  "X",
			want1: "AA0",
		},
		{
			name: "3",
			args: args{
				component: "XAAA0",
			},
			want:  "X",
			want1: "AAA0",
		},
		{
			name: "4",
			args: args{
				component: "XAAAA0",
			},
			want:  "X",
			want1: "AAAA0",
		},
		{
			name: "5",
			args: args{
				component: "XAAAAA0",
			},
			want:  "X",
			want1: "AAAAA0",
		},
		{
			"6",
			args{
				component: "K01",
			},
			"",
			"",
		},
		{
			"7",
			args{
				component: "x01",
			},
			"X",
			"01",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := splitComponentName(tt.args.component)
			if got != tt.want {
				t.Errorf("splitComponetName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("splitComponetName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
