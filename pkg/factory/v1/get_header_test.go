package factory

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getHeader(t *testing.T) {
	assert := assert.New(t)

	header := make(http.Header)
	header.Add("X-Token", "test")

	type args struct {
		configRaw map[string]interface{}
		lastOuput string
		inputs    []interface{}
	}
	tests := []struct {
		name    string
		header  http.Header
		args    args
		want    string
		wantErr bool
	}{
		{"no config will fail", nil, args{map[string]interface{}{"values": []int{1}}, "", nil}, "", true},
		{"no values", nil, args{map[string]interface{}{}, "", []interface{}{header}}, "", false},
		{"no inputs", nil, args{map[string]interface{}{}, "", nil}, "", true},
		{"valid config but no header set", nil, args{map[string]interface{}{"name": "X-Token"}, "", []interface{}{make(http.Header)}}, "", false},
		{"valid config and header set", nil, args{map[string]interface{}{"name": "X-Token"}, "", []interface{}{header}}, "test", false},
		{"invalid config name and no header set", nil, args{map[string]interface{}{"invalid": false}, "", []interface{}{make(http.Header)}}, "", false},
		{"invalid config name but header set", nil, args{map[string]interface{}{"invalid": false}, "", []interface{}{header}}, "", false},
		{"invalid config map", nil, args{map[string]interface{}{"name": []int{1}}, "", []interface{}{make(http.Header)}}, "", true},
		{"invalid input type", nil, args{map[string]interface{}{"invalid": false}, "", []interface{}{1}}, "", true},
	}
	for _, test := range tests {
		got, err := getHeader(test.args.configRaw, test.args.lastOuput, test.args.inputs...)
		assert.Equal(test.want, got, test.name)
		if test.wantErr {
			assert.Error(err, test.name)
		} else {
			assert.NoError(err, test.name)
		}
	}
}
