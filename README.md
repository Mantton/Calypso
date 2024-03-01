# Calypso

Calypso is a Swift/Typescript/Go inspired language created to help me learn more about programming languages and compilers.

# Data Types
- Signed Integers - `int`, `i64`, `i32`, `i16`, `i8`
- Unsigned Integers - `uint`, `u64`, `u32`, `u16`, `u8`
- Floating Points - `float`, `double `
- Booleans - `bool`
- Strings - `string`
- Characters - `char`
- Nil/Null - `nil`


# Syntax

## Modules

```
module main;
```

## Variables & Constants
```javascript

// immutable
const Foo = 20;

// mutable
let Bar = "hello, world.";
```

### Notes

Constants declared in the global scope must be known at compile time. for example
```swift
module main;

const Foo = 20; // Valid!
const Bar = Baz(); // Invalid!
```


## Functions
```swift
module main;


func main() {
    const Foo = 10;
}

func Foo() -> int {
    return 10;
}

// Function with Constrained Type Parameter
func Bar<T: Hashable>(a:T) -> T {
    return a;
}
```
