package airtable

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

const (
	testKeyVariable  string = "keyvariable"
	testBaseVariable string = "basevariable"
)

func Test_initializeClientWithoutKeyEnvVar(t *testing.T) {
	t.Log("testing InitializeClient: expect error, no AIRTABLE_KEY environment variable")
	_ = os.Unsetenv(airtableKeyVariable)
	_ = os.Setenv(airtableBaseVariable, testBaseVariable)

	err := testInitializeClient(
		t,
		defaultAirtableHost,
		testKeyVariable,
		testBaseVariable)

	if err == nil {
		t.Fatalf("Error should not be nil because the AIRTABLE_KEY environment variable is required to initialize the client")
	}

}
func Test_initializeClientWithoutBaseEnvVar(t *testing.T) {
	t.Log("testing InitializeClient: expect error, no AIRTABLE_BASE environment variable")
	_ = os.Setenv(airtableKeyVariable, testKeyVariable)
	_ = os.Unsetenv(airtableBaseVariable)

	err := testInitializeClient(
		t,
		defaultAirtableHost,
		testKeyVariable,
		testBaseVariable)

	if err == nil {
		t.Fatalf("Error should not be nil because the AIRTABLE_BASE environment variable is required to initialize the client")
	}
}

func Test_initializeClient(t *testing.T) {
	t.Log("testing InitializeClient: happy path")
	os.Setenv(airtableKeyVariable, testKeyVariable)
	os.Setenv(airtableBaseVariable, testBaseVariable)

	err := testInitializeClient(
		t,
		defaultAirtableHost,
		testKeyVariable,
		testBaseVariable)

	if err != nil {
		t.Fatalf("There was an error initializing the client successfully [%s]", err)
	}
}

func testInitializeClient(t *testing.T, atHost string, atKey string, atBase string) error {
	ac, err := InitializeClient()

	if err != nil {
		t.Logf("There was an error creating the client [%s]", err)
		return fmt.Errorf("There was an error creating the client [%s]", err)
	}

	if ac == nil {
		t.Logf("There was no error creating the client, but the client is nil %v", ac)
		return fmt.Errorf("There was no error creating the client, but the client is nil %v", ac)
	}

	t.Logf("Got Airtable Client")
	t.Logf("Airtable Key: '%s'", *ac.Key)
	t.Logf("Airtable URL: '%s'", *ac.URL)

	if *ac.Key != atKey {
		t.Errorf("The testKeyVariable [%s] was not set to the expected value [%s]",
			*ac.Key, atKey)
		return errors.New("airtableKeyVariable did not match")
	}

	testClientURL := fmt.Sprintf(atHost+"%s", atBase) + "/%s"

	if *ac.URL != testClientURL {
		t.Fatalf("The testClientURL [%s] was not set to the expected value [%s]",
			*ac.URL, testClientURL)
		return errors.New("airtableClientURL did not match")
	}

	return nil

}
