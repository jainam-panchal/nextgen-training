## Problem Statement

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

## Project Structure

- `cmd/garage/main.go` application entry point
- `internal/cli` contains command handling and output formatting
- `internal/garage` contains the parking logic, validation, models, and fee calculation

## Functional Requirements

Supports:

- parking a vehicle on a requested floor
- exiting a vehicle and calculating the fee
- finding a parked vehicle by plate number
- showing floor-wise availability status

## Edge Cases Handled

- invalid plate number
- invalid floor number
- missing command arguments
- duplicate vehicle entry
- floor full on requested floor
- searching for a vehicle that is not parked
- exiting a vehicle that is not parked

## Notes

- the vehicle is parked only on the requested floor
- the nearest spot means the first free spot on that floor
- billing is charged per started hour
- environment variables were added for flexibility, even though the base problem uses fixed defaults

## Env Variables

```env
PARKING_TOTAL_FLOORS=5
PARKING_SPOTS_PER_FLOOR=100
PARKING_HOURLY_RATE=20
```

## Commands

- run the application:
  - `go run ./cmd/garage`
- build the project:
  - `go build ./...`
- run tests:
  - `go test ./...`

## Alternate Approach

- Sorted free-list: Keep all free spots of a floor in sorted order. The nearest free spot is always the first one in the list. Parking is simple because we take the first free spot, and
  this can be treated as `O(1)` if we move the slice forward. Exit is slower because the freed spot has to be added back in the correct sorted position, which is `O(k)` in the worst case
  because elements may need to shift.
- Min-heap: Keep free spots in a heap where the smallest free spot is always on top. Parking removes the smallest spot in `O(log k)`, and exit adds the freed spot back in `O(log k)`. This
  gives better balanced performance for both park and exit.
 
Note: (k = number of free spots or spots on that floor)