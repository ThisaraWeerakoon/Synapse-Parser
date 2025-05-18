// main.go
package main

import (
	"fmt"
	"log"
	"poc_payload_processor/parser"
)

func main() {
	// Initialize the expression engine
	engine := parser.NewEngine()

	// --- XML Example ---
	fmt.Println("--- XML Example ---")
	xmlData := []byte(`<root><user><id>123</id><name>John Doe</name><active>true</active><balance>100.50</balance></user></root>`)
	xmlMsgCtx := parser.NewMessageContext(xmlData, "application/xml", engine)

	// Evaluate XPath expressions
	nameResult, err := xmlMsgCtx.EvaluateExpression("xpath:/root/user/name/text()")
	if err != nil {
		log.Fatalf("XML XPath Error: %v", err)
	}
	fmt.Printf("User Name (string): %s (Type: %s)\n", nameResult.Value, nameResult.Type)

	idResult, err := xmlMsgCtx.EvaluateExpression("xpath:/root/user/id") // Gets the node
	if err != nil {
		log.Fatalf("XML XPath Error: %v", err)
	}
	// To get text from a node, you might need a specific function or refine XPath
	// For antchfx/xpath, /text() is usually needed for text content.
	// If QueryResult.Value is an antchfx.NodeNavigator, you'd call String() on it.
	// For simplicity, our QueryResult.Value for XPath text() returns string directly.
	fmt.Printf("User ID Node (raw): %v (Type: %s)\n", idResult.Value, idResult.Type)
	// If we want the text value of a node directly from XPath that doesn't use text():
	// This depends on how the XMLPayload's Query is implemented.
	// For this PoC, `xpath:/root/user/id` might return a node object.
	// Let's assume our XPath engine is smart enough or we use text().

	activeResult, err := xmlMsgCtx.EvaluateExpression("xpath:boolean(/root/user/active='true')")
	if err != nil {
		log.Fatalf("XML XPath Error: %v", err)
	}
	fmt.Printf("User Active (bool): %t (Type: %s)\n", activeResult.Value, activeResult.Type)

	balanceResult, err := xmlMsgCtx.EvaluateExpression("xpath:number(/root/user/balance)")
	if err != nil {
		log.Fatalf("XML XPath Error: %v", err)
	}
	fmt.Printf("User Balance (float64): %f (Type: %s)\n", balanceResult.Value, balanceResult.Type)

	// --- JSON Example ---
	fmt.Println("\n--- JSON Example ---")
	jsonData := []byte(`{"store":{"book":[{"category":"reference","author":"Nigel Rees","title":"Sayings of the Century","price":8.95},{"category":"fiction","author":"Evelyn Waugh","title":"Sword of Honour","price":12.99}],"bicycle":{"color":"red","price":19.95}},"expensive":10}`)
	jsonMsgCtx := parser.NewMessageContext(jsonData, "application/json", engine)

	// Evaluate JSONPath expressions
	authorResult, err := jsonMsgCtx.EvaluateExpression("jsonpath:store.book.0.author")
	if err != nil {
		log.Fatalf("JSONPath Error: %v", err)
	}
	fmt.Printf("First book author (string): %s (Type: %s)\n", authorResult.Value, authorResult.Type)

	priceResult, err := jsonMsgCtx.EvaluateExpression("jsonpath:store.bicycle.price")
	if err != nil {
		log.Fatalf("JSONPath Error: %v", err)
	}
	fmt.Printf("Bicycle price (float64): %f (Type: %s)\n", priceResult.Value, priceResult.Type)

	// Example of getting an array
	allBookAuthors, err := jsonMsgCtx.EvaluateExpression("jsonpath:store.book.#.author") // gjson path for array of authors
	if err != nil {
		log.Fatalf("JSONPath Error: %v", err)
	}
	fmt.Printf("All book authors (slice): %v (Type: %s)\n", allBookAuthors.Value, allBookAuthors.Type)


	// --- Mixed Content Example (XML containing JSON) ---
	fmt.Println("\n--- Mixed Content Example (XML with embedded JSON) ---")
	mixedXMLData := []byte(`<order><id>789</id><customerName>Jane Doe</customerName><details><![CDATA[{"item": "laptop", "quantity": 1, "specs": {"ram": "16GB", "ssd": "512GB"}}]]></details></order>`)
	mixedXMLMsgCtx := parser.NewMessageContext(mixedXMLData, "application/xml", engine)

	// Extract JSON string from XML, then query JSON
	// Note: The pipe `|` needs to be URL-encoded if part of a URL, but fine in a string literal.
	// We'll use ` | ` (with spaces) as a delimiter for simplicity in PoC parsing.
	itemNameResult, err := mixedXMLMsgCtx.EvaluateExpression("xpath:/order/details/text() | extractAsJSON | jsonpath:item")
	if err != nil {
		log.Fatalf("Mixed Content Error: %v", err)
	}
	fmt.Printf("Item from embedded JSON (string): %s (Type: %s)\n", itemNameResult.Value, itemNameResult.Type)

	ramSpecResult, err := mixedXMLMsgCtx.EvaluateExpression("xpath:/order/details/text() | extractAsJSON | jsonpath:specs.ram")
	if err != nil {
		log.Fatalf("Mixed Content Error: %v", err)
	}
	fmt.Printf("RAM from embedded JSON (string): %s (Type: %s)\n", ramSpecResult.Value, ramSpecResult.Type)


	// --- Mixed Content Example (JSON containing XML) ---
	fmt.Println("\n--- Mixed Content Example (JSON with embedded XML) ---")
	mixedJSONData := []byte(`{"transactionId": "tx123", "productInfo": "<product><name>Super Widget</name><price>99.99</price></product>"}`)
	mixedJSONMsgCtx := parser.NewMessageContext(mixedJSONData, "application/json", engine)

	productNameResult, err := mixedJSONMsgCtx.EvaluateExpression("jsonpath:productInfo | extractAsXML | xpath:/product/name/text()")
	if err != nil {
		log.Fatalf("Mixed Content Error: %v", err)
	}
	fmt.Printf("Product name from embedded XML (string): %s (Type: %s)\n", productNameResult.Value, productNameResult.Type)

	// --- Error Handling Example: Invalid Path ---
	fmt.Println("\n--- Error Handling Example ---")
	_, err = jsonMsgCtx.EvaluateExpression("jsonpath:store.book.10.author") // Index out of bounds
	if err != nil {
		fmt.Printf("Expected error for invalid path: %v\n", err)
	}

	_, err = xmlMsgCtx.EvaluateExpression("xpath:/root/nonexistent/text()")
	if err != nil {
		fmt.Printf("Expected error for invalid XPath: %v\n", err)
	}

	_, err = mixedXMLMsgCtx.EvaluateExpression("xpath:/order/details/text() | extractAsJSON | jsonpath:nonexistent.key")
	if err != nil {
		fmt.Printf("Expected error for invalid path in embedded JSON: %v\n", err)
	}

	// --- Demonstrate Lazy Parsing ---
	fmt.Println("\n--- Lazy Parsing Demo ---")
	lazyData := []byte(`{"key": "lazyValue"}`)
	lazyMsgCtx := parser.NewMessageContext(lazyData, "application/json", engine)
	// At this point, lazyData is not yet parsed by JSONPayload.
	// Accessing processedPayload directly here would be for demonstration only,
	// normally it's an internal detail.
	// We can infer it by checking if EvaluateExpression works.
	fmt.Println("MessageContext created, payload not yet parsed (internally).")
	lazyResult, err := lazyMsgCtx.EvaluateExpression("jsonpath:key")
	if err != nil {
		log.Fatalf("Lazy parsing error: %v", err)
	}
	fmt.Printf("Lazy evaluated result: %s. Payload was parsed on first EvaluateExpression call.\n", lazyResult.Value)

	// Second call uses cached parsed payload
	lazyResult2, err := lazyMsgCtx.EvaluateExpression("jsonpath:key")
	if err != nil {
		log.Fatalf("Lazy parsing error on second call: %v", err)
	}
	fmt.Printf("Second call result: %s. Used cached parsed payload.\n", lazyResult2.Value)
}
