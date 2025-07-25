package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	azIdentity "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	armResources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	armStorage "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

var (
	location           = "westus"
	resourceGroupName  = "sample-resource-group"
	storageAccountName = fmt.Sprintf("samplestor%d", rand.Intn(1000))
)

var (
	resourcesClientFactory *armResources.ClientFactory
	storageClientFactory   *armStorage.ClientFactory
)

var (
	resourceGroupClient *armResources.ResourceGroupsClient
	accountsClient      *armStorage.AccountsClient
)

func main() {
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")

	// Get Azure credential
	cred, err := azIdentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
	}

	// Create background context to correlate async operations
	ctx := context.Background()

	// Instantiate a new ARM Client Factory
	resourcesClientFactory, err = armResources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Instantiate a new ARM Client
	resourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()

	log.Println("Creating Resource Group:", resourceGroupName)
	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armResources.ResourceGroup{
			Location: to.Ptr(location),
		},
		nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Resource Group Created:", *resourceGroupResp.ResourceGroup.ID)

	// Instantiate Azure Storage Client Factory
	storageClientFactory, err = armStorage.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Instantiate a new Azure Storage Client
	accountsClient = storageClientFactory.NewAccountsClient()

	// Start async operation
	log.Println("Create Storage Account:", storageAccountName)
	storageAcctOperation, err := accountsClient.BeginCreate(
		ctx,
		resourceGroupName,
		storageAccountName,
		armStorage.AccountCreateParameters{
			Kind:     to.Ptr(armStorage.KindStorageV2),
			Location: to.Ptr(location),
			SKU: &armStorage.SKU{
				Name: to.Ptr(armStorage.SKUNameStandardLRS),
				Tier: to.Ptr(armStorage.SKUTierStandard),
			},
		}, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Await async operation
	storageAcctResp, err := storageAcctOperation.PollUntilDone(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Storage Account Created:", *storageAcctResp.Account.ID)
}
