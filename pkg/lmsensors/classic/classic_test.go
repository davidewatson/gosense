package classic_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"experimental/dwat/gosense/pkg/cache"
	"experimental/dwat/gosense/pkg/lmsensors/classic"
	"experimental/dwat/gosense/pkg/report"
)

// classicFormatTests ensures real world hardware looks the same with new
// code. To generate new wants for wedge100:
//
// $ bmc=$(randbox asset_tagging.openbmc.wedge100)
// $ curl -s -k https://$bmc:8443/api/sys/sensors | pastebin
// https://phabricator.intern.facebook.com/P110356221
//
// To generate new command output from the same machine:
//
// $ ssh root@$bmc sensors | pastebin
// https://phabricator.intern.facebook.com/P110358789
//
// *Note* you will likely have to fix up the temperature values in order for
// the tests to pass. Also, the test cases are purposely []byte(s) to make it
// easier for developers to add test cases.
var classicFormatTests = []struct {
	name          string
	serviceOutput []byte
	sensorsOutput []byte
}{
	{
		name:          "wedge100",
		serviceOutput: []byte(`{"Information": [{"Adapter": "ast_i2c.3", "Outlet Middle Temp": "+26.5 C", "name": "tmp75-i2c-3-48"}, {"Adapter": "ast_i2c.3", "Inlet Middle Temp": "+22.6 C", "name": "tmp75-i2c-3-49"}, {"Adapter": "ast_i2c.3", "Inlet Left Temp": "+22.2 C", "name": "tmp75-i2c-3-4a"}, {"Adapter": "ast_i2c.3", "Switch Temp": "+37.3 C", "name": "tmp75-i2c-3-4b"}, {"Adapter": "ast_i2c.3", "Inlet Right Temp": "+24.0 C", "name": "tmp75-i2c-3-4c"}, {"+5V Voltage": "+5.06 V", "CPU Temp": "+48.0 C", "+12V Voltage": "+12.37 V", "+3V Voltage": "+3.28 V", "Adapter": "ast_i2c.4", "VDIMM Voltage": "+1.21 V", "CPU Vcore": "+1.80 V", "name": "com_e_driver-i2c-4-33", "Memory Temp": "+33.5 C"}, {"vout1": "+12.45 V", "Adapter": "ast_i2c.7", "iout1": "+9.58 A", "name": "ltc4151-i2c-7-6f"}, {"Fan 2 rear": "4800 RPM", "Fan 5 front": "7500 RPM", "Fan 3 rear": "4950 RPM", "Adapter": "ast_i2c.8", "name": "fancpld-i2c-8-33", "Fan 1 rear": "4950 RPM", "Fan 1 front": "7500 RPM", "Fan 2 front": "7500 RPM", "Fan 4 front": "7500 RPM", "Fan 4 rear": "4800 RPM", "Fan 3 front": "7500 RPM", "Fan 5 rear": "4950 RPM"}, {"Adapter": "ast_i2c.8", "name": "tmp75-i2c-8-48", "Outlet Right Temp": "+23.4 C"}, {"Outlet Left Temp": "+22.0 C", "Adapter": "ast_i2c.8", "name": "tmp75-i2c-8-49"}], "Resources": [], "Actions": []}`),
		sensorsOutput: []byte(`tmp75-i2c-3-48
Adapter: ast_i2c.3
Outlet Middle Temp:  +26.5 C  (high = +80.0 C, hyst = +75.0 C)

tmp75-i2c-3-49
Adapter: ast_i2c.3
Inlet Middle Temp:  +22.6 C  (high = +80.0 C, hyst = +75.0 C)

tmp75-i2c-3-4a
Adapter: ast_i2c.3
Inlet Left Temp:  +22.2 C  (high = +80.0 C, hyst = +75.0 C)

tmp75-i2c-3-4b
Adapter: ast_i2c.3
Switch Temp:  +37.3 C  (high = +80.0 C, hyst = +75.0 C)

tmp75-i2c-3-4c
Adapter: ast_i2c.3
Inlet Right Temp:  +24.0 C  (high = +80.0 C, hyst = +75.0 C)

com_e_driver-i2c-4-33
Adapter: ast_i2c.4
CPU Vcore:      +1.80 V
+3V Voltage:    +3.28 V
+5V Voltage:    +5.06 V
+12V Voltage:  +12.37 V
VDIMM Voltage:  +1.21 V
Memory Temp:    +33.5 C
CPU Temp:       +48.0 C

ltc4151-i2c-7-6f
Adapter: ast_i2c.7
vout1:       +12.45 V
iout1:        +9.58 A

fancpld-i2c-8-33
Adapter: ast_i2c.8
Fan 1 front: 7500 RPM
Fan 1 rear:  4950 RPM
Fan 2 front: 7500 RPM
Fan 2 rear:  4800 RPM
Fan 3 front: 7500 RPM
Fan 3 rear:  4950 RPM
Fan 4 front: 7500 RPM
Fan 4 rear:  4800 RPM
Fan 5 front: 7500 RPM
Fan 5 rear:  4950 RPM

tmp75-i2c-8-48
Adapter: ast_i2c.8
Outlet Right Temp:  +23.4 C  (high = +80.0 C, hyst = +75.0 C)

tmp75-i2c-8-49
Adapter: ast_i2c.8
Outlet Left Temp:  +22.0 C  (high = +80.0 C, hyst = +75.0 C)`),
	},
}

func TestClassicFormat(t *testing.T) {
	var observed, expected report.ClassicReport

	for _, tt := range classicFormatTests {
		err := json.Unmarshal(tt.serviceOutput, &expected)
		if err != nil {
			t.Fatalf("Failed to unmarshal expected JSON %v\n", err)
		}

		text := classic.Format(tt.sensorsOutput)
		err = json.Unmarshal(text, &observed)
		if err != nil {
			t.Fatalf("Failed to unmarshal observed JSON %v\n", err)
		}

		if !reflect.DeepEqual(expected, observed) {
			t.Fatalf("Classic format for %s does not match, observed \n%s\n, expected \n%s\n", tt.name, observed, expected)
		}
	}
}

// Benchmark profiles Get and UpdateWithTimeout.
func Benchmark(b *testing.B) {
	c := cache.NewCache("csensor", classic.Update, 60*60*24*365)
	b.ResetTimer()

	b.Run("c.Get()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c.Get()
		}
	})

	b.Run("c.UpdateWithTimeout()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = c.UpdateWithTimeout(false)
		}
	})
}
