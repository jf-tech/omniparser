package customfuncs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEval(t *testing.T) {
	for _, test := range []struct {
		name string
		expr string
		args []string
		err  string
		out  string
	}{
		{
			name: "empty expr",
			expr: "",
			args: nil,
			err:  "Unexpected end of expression",
			out:  "",
		},
		{
			name: "do nothing expr",
			expr: "()",
			args: nil,
			err:  "",
			out:  "",
		},
		{
			name: "int",
			expr: "1+2*3-4/(5-6)",
			args: nil,
			err:  "",
			out:  "11",
		},
		{
			name: "float",
			expr: "3.1415926 * 3.1415926",
			args: nil,
			err:  "",
			out:  "9.86960406437476",
		},
		{
			name: "bool",
			expr: "42 == 0",
			args: nil,
			err:  "",
			out:  "false",
		},
		{
			name: "string",
			expr: "('abcdefg' + '123456')",
			args: nil,
			err:  "",
			out:  "abcdefg123456",
		},
		{
			name: "args # not 2n",
			expr: "()",
			args: []string{"b"},
			err:  "invalid number of args to 'eval'",
			out:  "",
		},
		{
			name: "arg decl has no :",
			expr: "()",
			args: []string{"b", "1"},
			err:  "arg decl must be in format of '<arg_name>:<arg_type>'",
			out:  "",
		},
		{
			name: "arg decl has empty/whitespace name",
			expr: "()",
			args: []string{"    :int", "1"},
			err:  "arg_name in '<arg_name>:<arg_type>' cannot be a blank string",
			out:  "",
		},
		{
			name: "arg decl has not supported type",
			expr: "()",
			args: []string{"arg:json", "{}"},
			err:  "arg_type 'json' in '<arg_name>:<arg_type>' is not supported",
			out:  "",
		},
		{
			name: "arg value is not float",
			expr: "()",
			args: []string{"arg:float", "not a float"},
			err:  `strconv.ParseFloat: parsing "not a float": invalid syntax`,
			out:  "",
		},
		{
			name: "arg value is not int",
			expr: "()",
			args: []string{"arg:int", "not a int"},
			err:  `strconv.ParseFloat: parsing "not a int": invalid syntax`,
			out:  "",
		},
		{
			name: "arg value is not boolean",
			expr: "()",
			args: []string{"arg:boolean", "not a boolean"},
			err:  `strconv.ParseBool: parsing "not a boolean": invalid syntax`,
			out:  "",
		},
		{
			name: "missing arg",
			expr: "[abc] + 1",
			args: []string{},
			err:  "No parameter 'abc' found.",
			out:  "",
		},
		{
			name: "lbs -> kg",
			expr: "[lbs] * 0.453592",
			args: []string{"lbs:int", "225"},
			err:  "",
			out:  "102.0582",
		},
		{
			name: "add mr. prefix",
			expr: "'Mr. ' + [name]",
			args: []string{"name:string", "John"},
			err:  "",
			out:  "Mr. John",
		},
		{
			name: "F -> C",
			expr: "([fahrenheit] - 32) * 5 / 9",
			args: []string{"fahrenheit:float", "78"},
			err:  "",
			out:  "25.555555555555557",
		},
		{
			name: "1 cockroach == how many tons of TNT",
			expr: "[cockroach_mass_in_kg] * ([light_speed_in_ms] ** 2) / [ton_tnt_joule]",
			args: []string{
				"cockroach_mass_in_kg:float", "0.03",
				"light_speed_in_ms:int", "299792458",
				"ton_tnt_joule:float", "4.184e+9",
			},
			err: "",
			out: "644422.9293046015",
		},
		{
			name: "mood forecast",
			expr: "[weather_is_sunny] ? ':)' : ([raining] ? ':(' : 'meh')",
			args: []string{"weather_is_sunny:boolean", "false", "raining:boolean", "false"},
			err:  "",
			out:  "meh",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			out, err := eval(nil, test.expr, test.args...)
			if test.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, test.out, out)
			} else {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
				assert.Equal(t, "", out)
			}
		})
	}
}
