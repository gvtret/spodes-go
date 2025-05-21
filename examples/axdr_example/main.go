package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gvtret/spodes-go/pkg/axdr"
)

// MyData is a sample struct to demonstrate A-XDR encoding/decoding.
type MyData struct {
	ID        int32          // Changed to int32 for direct mapping to double-long
	Value     string         // Maps to visible-string
	Timestamp axdr.DateTime  // Using the library's DateTime type
	Active    bool           // Maps to boolean
	Readings  axdr.Array     // Array of uint16 readings
}

func main() {
	// 1. Create sample data
	originalData := MyData{
		ID:    12345,
		Value: "Sample A-XDR Data",
		Timestamp: axdr.FromTime(time.Date(2024, time.March, 15, 10, 30, 0, 0, time.UTC), false),
		Active:    true,
		Readings:  axdr.Array{uint16(100), uint16(150), uint16(200)},
	}
	fmt.Printf("Original Data: %+v\n\n", originalData)

	// To encode a struct like MyData, it needs to be represented as an axdr.Structure
	// or the individual fields need to be encoded sequentially if you're building a custom structure.
	// For simplicity, we'll represent MyData as an axdr.Structure.
	// The order of fields in the axdr.Structure matters for decoding.
	axdrStructureToEncode := axdr.Structure{
		originalData.ID,        // int32
		originalData.Value,     // string
		originalData.Timestamp, // axdr.DateTime
		originalData.Active,    // bool
		originalData.Readings,  // axdr.Array
	}
	fmt.Printf("Data prepared as axdr.Structure for encoding: %+v\n\n", axdrStructureToEncode)

	// 2. Encode the axdr.Structure
	encodedBytes, err := axdr.Encode(axdrStructureToEncode)
	if err != nil {
		log.Fatalf("Error encoding A-XDR: %v", err)
	}
	fmt.Printf("Encoded A-XDR Bytes: %X\n\n", encodedBytes)

	// 3. Decode the A-XDR bytes
	decodedInterface, err := axdr.Decode(encodedBytes)
	if err != nil {
		log.Fatalf("Error decoding A-XDR: %v", err)
	}
	fmt.Printf("Decoded Interface: %+v (Type: %T)\n\n", decodedInterface, decodedInterface)

	// 4. Type assert and verify
	// The decoded interface will be an axdr.Structure
	decodedStructure, ok := decodedInterface.(axdr.Structure)
	if !ok {
		log.Fatalf("Decoded interface is not an axdr.Structure, but %T", decodedInterface)
	}

	if len(decodedStructure) != 5 {
		log.Fatalf("Decoded structure has incorrect number of fields: got %d, want 5", len(decodedStructure))
	}

	// Extract and type assert individual fields
	// This requires knowing the order and types.
	decodedID, idOk := decodedStructure[0].(int32)
	decodedValue, valOk := decodedStructure[1].(string)
	decodedTimestamp, tsOk := decodedStructure[2].(axdr.DateTime)
	decodedActive, activeOk := decodedStructure[3].(bool)
	decodedReadingsInterface, readingsArrOk := decodedStructure[4].(axdr.Array)

	if !(idOk && valOk && tsOk && activeOk && readingsArrOk) {
		log.Fatalf("Error during type assertion of decoded fields.")
	}
	
	// Further convert readings if necessary
	var actualReadings axdr.Array
	for _, r := range decodedReadingsInterface {
		if val, rOk := r.(uint16); rOk {
			actualReadings = append(actualReadings, val)
		} else {
			log.Fatalf("Failed to assert reading element to uint16: got %T", r)
		}
	}


	// Construct the decoded MyData struct for comparison (optional, for verification)
	recoveredData := MyData{
		ID:        decodedID,
		Value:     decodedValue,
		Timestamp: decodedTimestamp,
		Active:    decodedActive,
		Readings:  actualReadings,
	}
	fmt.Printf("Recovered Data: %+v\n\n", recoveredData)

	// Verification (simple check)
	if recoveredData.ID == originalData.ID &&
		recoveredData.Value == originalData.Value &&
		// axdr.DateTime does not have a direct .Equal method, compare fields or use ToTime()
		recoveredData.Timestamp.Date.Year == originalData.Timestamp.Date.Year &&
		recoveredData.Timestamp.Time.Hour == originalData.Timestamp.Time.Hour &&
		recoveredData.Active == originalData.Active &&
		len(recoveredData.Readings) == len(originalData.Readings) { // Basic check for readings length
		
		match := true
		for i, v := range originalData.Readings {
			if recVal, ok := recoveredData.Readings[i].(uint16); !ok || recVal != v.(uint16) {
				match = false
				break
			}
		}
		if match {
			fmt.Println("Successfully encoded and decoded data. Original and recovered data match (basic check).")
		} else {
			fmt.Println("Data mismatch in Readings array.")
		}
	} else {
		fmt.Println("Data mismatch between original and recovered data.")
	}
}
