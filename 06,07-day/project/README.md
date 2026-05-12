### Requirements:
Build a ride-sharing dispatch system in Go:
1. DATA MODELS:
- Driver: { ID, Name, Location{Lat,Lng},
  Status(available/busy/offline),
  Rating float64, RideHistory []RideID }
- Rider: { ID, Name, Location{Lat,Lng}, PaymentMethod }
- Ride: { ID, RiderID, DriverID, Pickup, Dropoff, Status,
  RequestTime, StartTime, EndTime, Fare float64 }
2. CORE FEATURES:
   a) Driver Management:
- Register, go online/offline, update location
- Store in HashMap by ID, also index by zone (grid-based)b) Ride Request:
- Rider requests ride → enters priority queue (priority = wait time)
- System finds nearest available driver (Euclidean distance)
- Assign driver → move to active rides (linked list)
  c) Ride Completion:
- Complete ride → remove from active, add to history (both driver &
  rider)
- Calculate fare: base ₹50 + ₹12/km
- Update driver status back to available
  d) Dispatch Logic:
- Process ride queue: match oldest request with nearest driver
- If no driver available within 5km → keep in queue
- If request older than 10 min → cancel with notification
3. QUERIES:
- Find N nearest drivers to a location
- Driver earnings for today/this week
- Average wait time for riders
- Busiest zones (most ride requests)
4. GO REQUIREMENTS:
- Define interfaces: DriverStore, RideQueue, RideTracker
- Implement min-heap for priority queue (from scratch)
- Each data structure in its own package
- Comprehensive error handling
- Unit tests for each component
- CLI interface for all operations
5. PROJECT STRUCTURE:
   ride-sharing/
   ├── cmd/
   │
   └── main.go
# CLI entry point
├── internal/
│├── models/# Data models
│├── driver/# Driver store (HashMap-based)
│├── queue/# Priority queue (heap-based)
│├── rides/# Active ride tracker (linked list)
│└── dispatch/# Dispatch logic (composes all)├── go.mod
└── README.md       