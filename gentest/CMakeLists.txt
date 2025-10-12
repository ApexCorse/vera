cmake_minimum_required(VERSION 3.10)
project(gentest)

add_executable(test test.c vera.c unity/unity.c)
target_include_directories(test PUBLIC . unity)

