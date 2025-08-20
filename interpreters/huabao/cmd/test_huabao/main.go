package main

import (
    "fmt"
    "huabaoprotocol/features/huabao_protocol"
    "huabaoprotocol/features/jono"
    "encoding/json"
    "flag"
    "os"
)

func main() {
    // Define command line flags
    testDvrFlag := flag.Bool("test-dvr", false, "Test DVR format parsing with sample data")
    dvrDataFlag := flag.String("data", "", "Custom DVR data string to parse")
    flag.Parse()
    
    var testData string
    if *dvrDataFlag != "" {
        testData = *dvrDataFlag
    } else {
        // Sample DVR data with known coordinates
        testData = "$$dc0174,30,V114,0370703,,250613 091038,A0008,-99,9,146819999,19,37,274686000,12.00,7800,0000000000009383,0000000000000000,0.00,0.00,0.00,-2040896705,0.00,0,0|0.00|0|0|0|0|0|0|2258,1#"
    }
    
    if *testDvrFlag || *dvrDataFlag != "" {
        fmt.Println("Testing DVR format parsing with data:")
        fmt.Println(testData)
        fmt.Println("\n--- Original Algorithm Expected Values ---")
        fmt.Println("IMEI: 0370703")
        fmt.Println("Event: V114")
        fmt.Println("Latitude: 19.624475")
        fmt.Println("Longitude: -99.153408")
        fmt.Println("Speed: 12.00")
        fmt.Println("Azimuth: 78")
        
        // Parse with huabao_protocol
        huabaoJson, err := huabao_protocol.Parse(testData)
        if err != nil {
            fmt.Printf("Error parsing DVR data: %v\n", err)
            os.Exit(1)
        }
        
        // Pretty print the Huabao JSON
        var huabaoData map[string]interface{}
        if err := json.Unmarshal([]byte(huabaoJson), &huabaoData); err != nil {
            fmt.Printf("Error unmarshaling Huabao JSON: %v\n", err)
        } else {
            prettyJson, _ := json.MarshalIndent(huabaoData, "", "  ")
            fmt.Println("\n--- Huabao Protocol Output ---")
            fmt.Println(string(prettyJson))
        }
        
        // Convert to Jono protocol
        jonoJson, err := jono.Initialize(huabaoJson)
        if err != nil {
            fmt.Printf("Error converting to Jono protocol: %v\n", err)
            os.Exit(1)
        }
        
        // Pretty print the Jono JSON
        var jonoData map[string]interface{}
        if err := json.Unmarshal([]byte(jonoJson), &jonoData); err != nil {
            fmt.Printf("Error unmarshaling Jono JSON: %v\n", err)
        } else {
            prettyJson, _ := json.MarshalIndent(jonoData, "", "  ")
            fmt.Println("\n--- Jono Protocol Output ---")
            fmt.Println(string(prettyJson))
        }
    } else {
        fmt.Println("Please specify -test-dvr to run the test or -data with a custom DVR string")
    }
}
