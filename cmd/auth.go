package cmd

import (
	"fmt"

	"github.com/AshutoshPatole/ssm/internal/store"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Register or sign-in your user",
	Long: `The auth command provides functionality for user authentication.
It allows users to register new accounts or sign in to existing ones.
This command initializes the Firebase authentication service for secure user management.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := store.InitFirebaseOnce(); err != nil {
			logrus.Errorf("Failed to initialize Firebase: %s", err)
			PrintFirebaseSetupGuide()
			logrus.Fatal("Firebase initialization failed")
		}
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}

func PrintFirebaseSetupGuide() {
	fmt.Println("\nIt seems you haven't set up Firebase correctly. Please follow these detailed steps:")

	fmt.Println("\nStep 1: Set up Firebase Project")
	fmt.Println("  1. Go to the Firebase Console (https://console.firebase.google.com/)")
	fmt.Println("  2. Click 'Add project' or select an existing project")
	fmt.Println("  3. Follow the prompts to set up your project")
	fmt.Println("  4. Once the project is created, you'll be taken to the project dashboard")

	fmt.Println("\nStep 2: Configure Firebase Authentication")
	fmt.Println("  1. In the left sidebar, click 'Authentication'")
	fmt.Println("  2. Click 'Get started' if it's your first time, or 'Sign-in method' if already set up")
	fmt.Println("  3. Find 'Email/Password' in the list of sign-in providers")
	fmt.Println("  4. Click the toggle switch to enable it")
	fmt.Println("  5. Click 'Save' to confirm the changes")

	fmt.Println("\nStep 3: Set up Firestore Database")
	fmt.Println("  1. In the left sidebar, click 'Firestore Database'")
	fmt.Println("  2. Click 'Create database'")
	fmt.Println("  3. Choose 'Start in production mode' for better security")
	fmt.Println("  4. Select a location for your database (choose the closest to your location)")
	fmt.Println("  5. Click 'Enable' to create the database")

	fmt.Println("\nStep 4: Configure Firestore Security Rules")
	fmt.Println("  1. After creating the database, you'll be in the 'Data' tab")
	fmt.Println("  2. Click the 'Rules' tab at the top")
	fmt.Println("  3. Replace the default rules with more secure ones, for example:")
	fmt.Println("     rules_version = '2';")
	fmt.Println("     service cloud.firestore {")
	fmt.Println("       match /databases/{database}/documents {")
	fmt.Println("         match /{document=**} {")
	fmt.Println("           allow read, write: if request.auth != null;")
	fmt.Println("         }")
	fmt.Println("       }")
	fmt.Println("     }")
	fmt.Println("  4. Click 'Publish' to apply the new rules")

	fmt.Println("\nStep 5: Register a Web App and Obtain Firebase API Key")
	fmt.Println("  1. In the left sidebar, click the gear icon next to 'Project Overview'")
	fmt.Println("  2. Select 'Project settings'")
	fmt.Println("  3. In the 'General' tab, scroll down to 'Your apps' section")
	fmt.Println("  4. Click the '</>' icon to add a web app")
	fmt.Println("  5. Give your app a nickname and click 'Register app'")
	fmt.Println("  6. In the SDK setup, you'll see a firebaseConfig object")
	fmt.Println("  7. Copy the 'apiKey' value from this object")

	fmt.Println("\nStep 6: Set Up Service Account")
	fmt.Println("  1. In Project settings, click the 'Service Accounts' tab")
	fmt.Println("  2. Under 'Firebase Admin SDK', click 'Generate new private key'")
	fmt.Println("  3. Save the downloaded JSON file")

	fmt.Println("\nStep 7: Configure SSM")
	fmt.Println("  1. Open the SSM configuration file located at ~/.ssm.yaml")
	fmt.Println("  2. Add the following lines to the configuration file, replacing the placeholders:")
	fmt.Println("     firebaseConfig: /path/to/downloaded/service-account.json")
	fmt.Println("     firebaseApiKey: YOUR_FIREBASE_API_KEY")
	fmt.Println("  3. Save the configuration file")
	fmt.Println("\nMake sure to replace '/path/to/downloaded/service-account.json' with the actual path")
	fmt.Println("to your downloaded JSON file, and 'YOUR_FIREBASE_API_KEY' with the API key you obtained")
	fmt.Println("from the Firebase Console.")

	fmt.Println("\nAfter completing these steps, try running the command again.")
}
