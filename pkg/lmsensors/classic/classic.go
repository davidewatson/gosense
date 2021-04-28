package classic

import (
	"regexp"
	"strings"

	"experimental/dwat/gosense/pkg/cache"
	"experimental/dwat/gosense/pkg/report"
)

const (
	commandPath    = "sensors" // Relative path to command
	processTimeout = 10        // Number of seconds before we attempt to kill the process
)

// Update runs the command and renders its output in the classic format.
func Update() ([]byte, error) {
	output, err := cache.RunCommand(cache.Command{
		Command: commandPath,
		Timeout: processTimeout,
	})

	if err != nil {
		return []byte(nil), err
	}

	return Format(output), nil
}

// Format takes stdout from the sensors command and returns the classic format
// defined by the python implementation here:
// https://fburl.com/diffusion/scvlbt7e
func Format(stdout []byte) []byte {
	result := make([]map[string]string, 0)

	// Throw away anything inbetween parentheses.
	re := regexp.MustCompile(`\(.+?\)`)
	data := re.ReplaceAllString(string(stdout), ``)

	for _, edata := range strings.Split(data, "\n\n") {
		adata := strings.SplitN(edata, "\n", 2)
		sresult := make(map[string]string)
		if len(adata) < 2 {
			break
		}
		sresult["name"] = strings.TrimSpace(adata[0])
		for _, sdata := range strings.Split(adata[1], "\n") {
			tdata := strings.Split(sdata, ":")
			if len(tdata) < 2 {
				continue
			}
			sresult[strings.TrimSpace(tdata[0])] = strings.TrimSpace(tdata[1])
		}

		result = append(result, sresult)
	}

	return report.FormatClassicInformation(result)
}
