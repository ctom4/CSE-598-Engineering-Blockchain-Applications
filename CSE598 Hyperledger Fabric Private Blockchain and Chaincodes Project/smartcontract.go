// smartcontract.go
// Hyperledger Fabric Private Blockchain and Chaincodes Project
//
// Author: Casey Tom
// ASURITE ID: ctom4
// Description:
// This smart contract defines a supply chain management system using Hyperledger Fabric.
// It supports creating, updating, transferring, and querying product records on a private blockchain ledger.
// Each product has an ID, name, status, owner, timestamps, description, and category.
//
// Implemented Functions:
// - InitLedger: Initializes the ledger with sample products
// - CreateProduct: Adds a new product to the ledger
// - UpdateProduct: Updates an existing product’s status, owner, description, and category
// - TransferOwnership: Changes the ownership of a product
// - QueryProduct: Retrieves a product by ID
//
// Helper Methods:
// - storeProduct: Stores or updates product data on the ledger
// - ProductExists: Checks if a product with a given ID exists
// - GetAllProducts: Retrieves all products stored in the ledger
// - getCurrentTimestamp: Returns the transaction timestamp in RFC3339 format
package main // Package main. Do not change this line.

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Product represents a product tracked in the supply chain ledger.
type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Owner       string `json:"owner"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// SupplyChainContract provides functions for managing products.
type SupplyChainContract struct {
	contractapi.Contract
}

// getCurrentTimestamp returns the current transaction timestamp in RFC3339 format.
func (s *SupplyChainContract) getCurrentTimestamp(ctx contractapi.TransactionContextInterface) (string, error) {
	txTime, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return "", fmt.Errorf("failed to get transaction timestamp: %v", err)
	}
	return time.Unix(txTime.Seconds, int64(txTime.Nanos)).Format(time.RFC3339), nil
}

// InitLedger adds predefined sample products to the ledger.
func (s *SupplyChainContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	timestamp, err := s.getCurrentTimestamp(ctx)
	if err != nil {
		return err
	}

	initialProducts := []Product{
		{
			ID:          "p1",
			Name:        "Laptop",
			Status:      "Manufactured",
			Owner:       "CompanyA",
			CreatedAt:   timestamp,
			UpdatedAt:   timestamp,
			Description: "High-end gaming laptop",
			Category:    "Electronics",
		},
		{
			ID:          "p2",
			Name:        "Smartphone",
			Status:      "Manufactured",
			Owner:       "CompanyB",
			CreatedAt:   timestamp,
			UpdatedAt:   timestamp,
			Description: "Latest model smartphone",
			Category:    "Electronics",
		},
	}

	for _, product := range initialProducts {
		if err := s.storeProduct(ctx, &product); err != nil {
			return err
		}
	}

	return nil
}

// CreateProduct adds a new product to the ledger.
func (s *SupplyChainContract) CreateProduct(ctx contractapi.TransactionContextInterface, productID, name, owner, description, category string) error {
	exists, err := s.ProductExists(ctx, productID)
	if err != nil {
		return fmt.Errorf("failed to check if product exists: %v", err)
	}
	if exists {
		return fmt.Errorf("product with ID '%s' already exists", productID)
	}

	currentTime, err := s.getCurrentTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}

	newProduct := Product{
		ID:          productID,
		Name:        name,
		Status:      "Manufactured",
		Owner:       owner,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
		Description: description,
		Category:    category,
	}

	if err := s.storeProduct(ctx, &newProduct); err != nil {
		return fmt.Errorf("failed to store product: %v", err)
	}

	return nil
}

// UpdateProduct modifies an existing product's details.
func (s *SupplyChainContract) UpdateProduct(ctx contractapi.TransactionContextInterface, productID, status, newOwner, description, category string) error {
	productBytes, err := ctx.GetStub().GetState(productID)
	if err != nil {
		return fmt.Errorf("failed to retrieve product '%s': %v", productID, err)
	}
	if productBytes == nil {
		return fmt.Errorf("product with ID '%s' does not exist", productID)
	}

	var product Product
	if err := json.Unmarshal(productBytes, &product); err != nil {
		return fmt.Errorf("failed to unmarshal product data: %v", err)
	}

	product.Status = status
	if newOwner != "" {
		product.Owner = newOwner
	}
	product.Description = description
	product.Category = category

	product.UpdatedAt, err = s.getCurrentTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("failed to get timestamp: %v", err)
	}

	return s.storeProduct(ctx, &product)
}

// TransferOwnership changes the owner of a product.
func (s *SupplyChainContract) TransferOwnership(ctx contractapi.TransactionContextInterface, productID, newOwner string) error {
	productBytes, err := ctx.GetStub().GetState(productID)
	if err != nil {
		return fmt.Errorf("failed to retrieve product '%s': %v", productID, err)
	}
	if productBytes == nil {
		return fmt.Errorf("product with ID '%s' does not exist", productID)
	}

	var product Product
	if err := json.Unmarshal(productBytes, &product); err != nil {
		return fmt.Errorf("failed to unmarshal product data: %v", err)
	}

	product.Owner = newOwner
	product.UpdatedAt, err = s.getCurrentTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("failed to get timestamp: %v", err)
	}

	return s.storeProduct(ctx, &product)
}

// QueryProduct retrieves a product from the ledger by ID.
func (s *SupplyChainContract) QueryProduct(ctx contractapi.TransactionContextInterface, productID string) (*Product, error) {
	productBytes, err := ctx.GetStub().GetState(productID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve product '%s': %v", productID, err)
	}
	if productBytes == nil {
		return nil, fmt.Errorf("the product does not exist")
	}

	var product Product
	if err := json.Unmarshal(productBytes, &product); err != nil {
		return nil, fmt.Errorf("failed to unmarshal product data: %v", err)
	}

	return &product, nil
}

// GetAllProducts returns all products stored in the ledger.
func (s *SupplyChainContract) GetAllProducts(ctx contractapi.TransactionContextInterface) ([]*Product, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var products []*Product
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var product Product
		if err := json.Unmarshal(queryResponse.Value, &product); err != nil {
			return nil, err
		}

		products = append(products, &product)
	}

	return products, nil
}

// ProductExists checks if a product exists in the ledger.
func (s *SupplyChainContract) ProductExists(ctx contractapi.TransactionContextInterface, productID string) (bool, error) {
	productBytes, err := ctx.GetStub().GetState(productID)
	if err != nil {
		return false, fmt.Errorf("failed to read from ledger: %v", err)
	}
	return productBytes != nil, nil
}

// storeProduct is a helper method to store or update a product in the ledger.
func (s *SupplyChainContract) storeProduct(ctx contractapi.TransactionContextInterface, product *Product) error {
	productBytes, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to serialize product: %v", err)
	}
	return ctx.GetStub().PutState(product.ID, productBytes)
}

// main starts the chaincode in the Hyperledger Fabric runtime.
func main() {
	chaincode, err := contractapi.NewChaincode(&SupplyChainContract{})
	if err != nil {
		fmt.Printf("Error creating supply chain chaincode: %s\n", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting supply chain chaincode: %s\n", err.Error())
	}
}
