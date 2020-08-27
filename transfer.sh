#!/bin/bash

./block send  張三 李四 10 班長 "張三轉李四10"
./block send  張三 王五 20 班長 "張三轉王五20"
./checkBalance.sh

echo "======================"
./block send  王五 李四 2 班長 "王五轉李四2"
./block send  王五 李四 3 班長 "王五轉李四3"
./block send  王五 張三 5 班長 "王五轉張三5"
./checkBalance.sh

echo "======================"
./block send  李四 趙六 14 班長 "李四轉趙六14"
./checkBalance.sh

