# auto-scalar
auto scalar application for a rest API. 

At started off with following approach from the base document shared. 

1. Declare types for the AppStatus and Replicas
2. call /app/status to get CPU and replica count.
3. calculate the replica new count in a way that CPU <.80
   1. inc replica will dec CPU and dec replica will inc CPU.
4. call /app/replicas to update the replica count.
5. repeat step 2 to 4 until CPU as much as close to .80 for every given time interval.


For point 1, based on the given sample response and constraint data, I declared the constants and defined types for AppStatus and Replicas. </br> 
For point 2, Started with http package but moved to use resty package as it was easier to use and code is more readable. </br>
For point 3, replica is inversely proportional to CPU. calculated replicas count as factor of current replicas with current/target CPU. </br>
For point 4, used resty package to update the replica count with Replicas. </br>

Improvements:
   1. Don't update the same replica count - Done
   2. Refactor the code to make it more readable - Done
   3. Exit gracefully on exiting the program - Done
   4. Add more configuration options like env based config. - Done
   5. Add more logging and error handling. - Done
   6. Add unit tests - done
   7. update code docs. - done
   8. containerize the application. - done

To build the application, run the following command to generate the binary in the name of auto-scalar.
```
go build -o auto-scalar .
```

To run this application, run the following command. 
```
go run .
```

To run the unit tests, run the following command. 
```
go test . 
```

Container

To build the container, run the following command. 
```
docker build -t auto-scalar .
```

To Run the container, run the following command. 
```
docker run -e STATUS_API_URL=http://host.docker.internal:8123/app/status -e REPLICAS_API_URL=http://host.docker.internal:8123/app/replicas -e ENVIRONMENT=prod auto-scalar
```

Note: update the STATUS_API_URL and REPLICAS_API_URL with respective endpoints based on the platfrom. Given example is for the MacOS


Potentially this application can be further iterated to make it more organized and also depending on the deployment platform and service discovery. 