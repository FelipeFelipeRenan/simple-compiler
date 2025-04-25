define i32 @sum(i32 %a, i32 %b) {
entry:
  %t0 = alloca i32 
  store i32 %a, i32* %t0
  %t1 = alloca i32 
  store i32 %b, i32* %t1
  %t2 = load i32, i32* %t0
  %t3 = load i32, i32* %t1
  %t4 = add i32 %t2, %t3
  ret i32 %t4
}

define i32 @main() {
entry:
  %t5 = call i32 @sum(i32 2, i32 3)
  ret i32 %t5
}

