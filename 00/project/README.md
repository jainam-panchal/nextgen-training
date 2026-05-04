## Given Problem Statement

Build a CLI-based parking garage management system in Go:

1. STRUCTURE:- 500 parking spots across 5 floors (100 per floor)

- Each spot: ID, floor, occupied (bool), vehicle plate, entry time
- Use a fixed-size array: [500]ParkingSpot

2. CORE FEATURES:

- Vehicle entry: Assign nearest available spot on requested floor
- Vehicle exit: Free spot, calculate duration & fee (₹20/hour)
- Display: Floor-wise availability summary
- Search: Find vehicle by plate number

3. GO REQUIREMENTS:

- Define ParkingSpot struct with proper types
- Use time.Now() and time.Since() for duration
- Handle errors properly (spot not found, garage full, invalid floor)
- Accept user input via fmt.Scan or bufio.Scanner

4. OUTPUT FORMAT:
   Entry: "Vehicle MH-12-AB-1234 parked at Floor 3, Spot 247"
   Exit:
   "Vehicle MH-12-AB-1234 exited. Duration: 2h 15m. Fee: ₹60"
   Status: "Floor 1: 73/100 available | Floor 2:45/100 available | ..."
