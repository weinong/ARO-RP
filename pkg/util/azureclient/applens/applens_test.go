package applens

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Azure/ARO-RP/pkg/util/azureclient"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

func TestMe(t *testing.T) {

	cred, err := azidentity.NewClientSecretCredential(
		"<fill-this>",
		"<fill-this>",
		"<fill-this>", nil)

	if err != nil {
		t.Fatal("new cred failed", err)
	}

	client := NewAppLensClient(&azureclient.PublicCloud, cred)
	resp, err := client.ListDetectors(context.Background(), nil)
	if err != nil {
		t.Fatal("List failed", err)
	}
	body, err := json.Marshal(resp.Body)
	if err != nil {
		t.Fatal("Marshal failed", err)
	}

	t.Log(string(body))
}
