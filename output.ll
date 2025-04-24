declare i32 @printf(i8*, ...)
@.str  = private unnamed_addr constant [4 x i8] c"%d\0A\00", align 1

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
  %t5 = alloca i32 
  %t7 = sub i32 0, 10
  %t6 = call i32 @sum(i32 2, i32 %t7)
  store i32 %t6, i32* %t5
  %t8 = load i32, i32* %t5
  %t9 = getelementptr  [4 x i8], [4 x i8]* @.str, i32 0, i32 0
  %t10 = call i32 (i8*, ...) @printf(i8* %t9, i32 %t8)
  %t11 = load i32, i32* %t5
  ret i32 %t11
}

