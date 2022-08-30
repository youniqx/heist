package pki

import (
	"testing"

	"github.com/go-test/deep"
)

func TestStringArray_UnmarshalJSON(t *testing.T) {
	type args struct {
		b []byte
	}

	tests := []struct {
		name    string
		args    args
		want    StringArray
		wantErr bool
	}{
		{
			name: "should correctly unmarshal value",
			args: args{
				b: []byte(`["some","value"]`),
			},
			want:    StringArray{"some", "value"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StringArray{}
			if err := s.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := deep.Equal(s, tt.want); diff != nil {
				t.Errorf("UnmarshalJSON() diff %v", diff)
			}
		})
	}
}

func TestStringArray_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		s       StringArray
		want    string
		wantErr bool
	}{
		{
			name:    "should correctly marshal string",
			s:       StringArray{"some", "value"},
			want:    `"some,value"`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(string(got), tt.want); diff != nil {
				t.Errorf("MarshalJSON() diff %v", diff)
			}
		})
	}
}
