Build an undo/redo system using a doubly linked list in Go:

1. STRUCTURE:

- Implement DoublyLinkedList[T] from scratch (NO container/list)
- Node: { Data T, Prev *Node[T], Next *Node[T] }
- Each node represents an EditOperation

2. EDIT OPERATIONS (implement all):

- Insert text at position
- Delete text at position (with length)
- Replace text at position
- Format change (bold, italic — store as metadata)

3. CORE FEATURES:- Apply edit → append to list, set as current

- Undo → move current pointer backward, reverse the operation
- Redo → move current pointer forward, reapply the operation
- When new edit after undo → discard all "future" edits (branch point)
- Display current document state
- Show edit history with current position marker

4. GO REQUIREMENTS:

- Use Go generics for the linked list: type Node[T any] struct
- Define EditOperation as an interface with Apply() and Reverse()
  methods
- Use pointer receivers for all mutating methods
- Implement fmt.Stringer for the list (print history nicely)

5. EXAMPLE SESSION:
   > insert 0 "Hello"→ doc: "Hello"
   > insert 5 " World"→ doc: "Hello World"
   > undo→ doc: "Hello"
   > undo→ doc: ""
   > redo→ doc: "Hello"
   > insert 5 " Go!"→ doc: "Hello Go!" (discards " World" branch)
