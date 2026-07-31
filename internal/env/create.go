package env

import (
	"fmt"
	"os"
	"path/filepath"

	config "github.com/AaltoRSE/shark-tank/internal/config"
	utils "github.com/AaltoRSE/shark-tank/internal/utils"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/spf13/viper"
)

// CreateEnvironment creates a new environment configuration in the viper config.
func CreateEnvironment(name string, fakeHome string, projectMounts []string) error {

	var envs map[string]any

	// Determine if the environment already exists
	envs = viper.GetStringMap("envs")

	if envs[name] != nil {
		fmt.Println("Environment already exists.")
		return nil
	}

	// Validate the environment name
	if !utils.CheckEnvironmentName(name) {
		fmt.Println("Invalid environment name:", name)
		fmt.Println("Environment name must be non-empty and contain only alphanumeric characters and underscores.")
		return nil
	}

	// Get the absolute path of the fake home directory
	absFakeHome, err := filepath.Abs(fakeHome)
	if err != nil {
		fmt.Println("Error getting absolute path for fake home directory:", err)
		return err
	}

	// Validate that the fake home directory exists
	if !utils.CheckFolderExists(absFakeHome) {
		// Ask to create the fake home directory if it doesn't exist
		fmt.Println("Fake home directory does not exist:", absFakeHome)
		confirm := confirmation.New("Do you want to create the fake home directory?", confirmation.Undecided)
		ready, err := confirm.RunPrompt()
		if err != nil {
			fmt.Println("Error during confirmation prompt:", err)
			return err
		}
		if ready {
			err := os.MkdirAll(absFakeHome, 0755)
			if err != nil {
				fmt.Println("Failed to create fake home directory:", err)
				return err
			}
			fmt.Println("Fake home directory created:", absFakeHome)
		} else {
			fmt.Println("Aborting environment creation due to missing fake home directory.")
			return nil
		}
	}

	// Validate that all project mounts exist
	for i, mount := range projectMounts {
		absMount, err := filepath.Abs(mount)
		if err != nil {
			fmt.Println("Error getting absolute path for project mount:", mount, err)
			return err
		}
		if !utils.CheckFolderExists(absMount) {
			fmt.Println("Project mount does not exist:", absMount)
			return nil
		}
		projectMounts[i] = absMount
	}

	// Create environment configuration
	fmt.Println("Creating environment with name:", name)
	viper.Set("envs."+name, map[string]interface{}{
		"home":   absFakeHome,
		"mounts": projectMounts,
	})

	// Write the updated configuration back to the config file
	config.WriteConfig()

	fmt.Println("Environment created successfully.")
	return nil
}
