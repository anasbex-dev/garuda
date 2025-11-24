cmake_minimum_required(VERSION 3.10)
project(GarudaMC)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -O2 -pthread")

# Include directories
include_directories(include)

# Source files
file(GLOB SRC_FILES "src/*.cpp" "src/*/*.cpp")

add_executable(garudamc ${SRC_FILES})

# Libraries
target_link_libraries(garudamc pthread)