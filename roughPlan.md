


A Map Function 
A reduce Function

Worker pool - Fix max goroutines to three

# General Idea
Start Program
Limit max routines to three
While not done reading directory and routine is free, assign a new file to process
    => 
    - Job queue with jobs or files to be processed
    - Passing jobs to workers using channels
    - Map key(State), values(city, population) -> key(state), values(city, popultion based on min_pop condition)
    - Reduce key(State), values(city, population) -> sum(state)
        - Output result formated as:

        Texas: 3
        - Dallas, 1257676
        - San Antonio, 1409019
        - Houston, 2195914
        California: 2
        - Los Angeles, 3884307
        - San Diego, 1355896
        Arizona: 1
        - Phoenix, 1513367
        Illinois: 1
        - Chicago, 2718782
        New York: 1
        - New York, 8405837
        Pennsylvania: 1
        - Philadelphia, 1553165


# More polished work flow
- Input dir and min_pop from cli
- Start WorkerPool


- ## Work Container
- Has filename
- Has min_pop
- Has result channel

- ## Result Container
- Has state
- Has list of cities paired with population using go map




## Main
- Get cli args
- Create WorkerPool
- Read filenames and add to jobs queue which is a channel with work stuct containing filename, min_pop, and channel for results. This is done with go routine (GetFilesForProcessing)


- ## WorkerPool
- Create fixed number of goroutines (workers)
- Start Workers

- ### Worker(jobs queue)
- Search through each received filename
    - for line in file
    - Use "," delimeter or csv reading package
    - if pop >= min_pop
    - add to results container as city, pop
    - return result as state to city, pop let reduce handle the rest
    - Send state map
- Is Map function with file values


- ### GetFilesForProcessing (Go routine)

- ## Reduce key(State), values(city, population) -> sum(state) 
- For results in all results which is a channel of channels, sum state = len(state), loop to display city values



# Concepts
- Workers
- Channels
- Wait Groups
- Map
- Reduce
- Worker Pool