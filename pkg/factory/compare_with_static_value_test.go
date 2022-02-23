package factory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_compareWithStaticValue(t *testing.T) {
	assert := assert.New(t)

	type args struct {
		configRaw map[string]interface{}
		lastOuput string
		inputs    []interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"no config will fail", args{map[string]interface{}{"values": []int{1}}, "", nil}, "", true},
		{"no values", args{map[string]interface{}{}, "", nil}, "t", false},
		{"no values", args{map[string]interface{}{}, "test", nil}, "f", false},
		{"one value", args{map[string]interface{}{
			"value": "test",
		}, "test", nil}, "t", false},
		{"one value in list", args{map[string]interface{}{
			"values": []string{"test"},
		}, "test", nil}, "t", false},
		{"one value dont equals", args{map[string]interface{}{
			"value": "foo",
		}, "bar", nil}, "f", false},
		{"one value in list dont equals", args{map[string]interface{}{
			"values": []string{"foo"},
		}, "bar", nil}, "f", false},

		{"correct call with extra useless inputs", args{map[string]interface{}{
			"values": []string{"test"},
		}, "test", []interface{}{"extra_useless"}}, "", true},
	}
	for _, test := range tests {
		got, err := compareWithStaticValue(test.args.configRaw, test.args.lastOuput, test.args.inputs...)
		assert.Equal(test.want, got, test.name)
		if test.wantErr {
			assert.Error(err, test.name)
		} else {
			assert.NoError(err, test.name)
		}
	}
}
