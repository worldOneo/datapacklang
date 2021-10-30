# Data Pack Lang - High Level language for mcfunctions
This is a transpiler for an custom higher level language to minecraft commands
## Example code
```
create store someStore

someStore[someVar] = 100

if someStore[someVar] == 100 {
  'say counter reached 100'
  'say will be reseted now'
  someStore[someVar] = 2 - 2
  if someStore[someVar] == 0 {
    'say someVar reseted'
  }
}
```
Will be transpiled to the working commands:
```
scoreboard objectives add a dummy
scoreboard players set b a 100
scoreboard objectives add c dummy
scoreboard players operation f c = b a
scoreboard players set g c 100
execute if score f c = g c run say counter reached 100
execute if score f c = g c run say will be reseted now
execute if score f c = g c run scoreboard players set j c 2
execute if score f c = g c run scoreboard players set k c 2
execute if score f c = g c run scoreboard players operation j c -= k c
execute if score f c = g c run scoreboard players operation b a = j c
execute if score f c = g c run scoreboard players operation k c = b a
execute if score f c = g c run scoreboard players set j c 0
execute if score f c = g c run execute if score k c = j c run say someVar reseted
```

## Todo
  - [x] Variables
  - [x] Basic Calculations
  - [x] If-Statement
  - [x] Transpile datapacks recursively

## Code examples
### Store
Store is the basic type to handle data.  
Create:
```
create store myStore
```
Set
```
myStore[myValue] = 10
```
Update
```
myStore[myValue] -= 5
```
Read
```
myStore[anotherValue] = myStore[myValue]
```
### Calculations
Inline calculations are possible e.g:
```
myStore[test] = myOtherStore[test] + myStore[addThis] - myStore[subThis]
```

### If-Statements
```
if myStore[valueA] < yourStore[valueA] {
  'say you have more than I'
}
```
Not
```
if not myStore[valueA] == 0 {
  'say valueA is not 0'
}
```

### Own commands
Own commands must be declared with '
```
'say im a own command'
```

### As
To execute as use `as`:
```
as '@e[type=minecraft:villager]' {
  'say I am a villager'
}
```