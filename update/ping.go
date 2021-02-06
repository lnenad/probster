package update

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"

	gv "github.com/hashicorp/go-version"
	"github.com/lnenad/probster/communication"
)

const serverURL = "https://probster.com/"
const path = "version.json"

var failedAttempts = 0

type VersionResult struct {
	Windows string
	Ubuntu  string
	Mac     string
}

func CheckVersion(current *gv.Version) (bool, string) {
	_, response, err := communication.Send(serverURL+path, "GET", nil, "")
	if err != nil {
		failedAttempts++
	}

	var vr VersionResult
	err = json.Unmarshal(response, &vr)
	if err != nil {
		log.Printf("Error while unmarshaling version response: %s", err)
		failedAttempts++
	}

	var latest *gv.Version

	switch runtime.GOOS {
	case "windows":
		fmt.Println("Hello from Windows")
		latest, err = gv.NewVersion(vr.Windows)
		if err != nil {
			log.Printf("Invalid version response from server: %s", err)
		}
	case "linux":
		fmt.Println("Hello from linux")
		latest, err = gv.NewVersion(vr.Ubuntu)
		if err != nil {
			log.Printf("Invalid version response from server: %s", err)
		}
	case "darwin":
		fmt.Println("Hello from darwin")
		latest, err = gv.NewVersion(vr.Mac)
		if err != nil {
			log.Printf("Invalid version response from server: %s", err)
		}
	default:
		log.Printf("Error while comparing runtime GOOS. Invalid GOOS value: %s", runtime.GOOS)
		return false, ""
	}

	if latest == nil {
		return false, ""
	}

	return latest.GreaterThan(current), latest.String()
}
