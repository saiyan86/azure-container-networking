# Use an official golang runtime as a parent image
FROM golang:1.8

# Set the working directory to /npm
WORKDIR /npm

# Copy the current directory contents into the container at /npm
ADD . /npm

# Install any needed packages specified in requirements.txt
RUN go get -d -v ./...
RUN go build

# Make port 80 available to the world outside this container
#EXPOSE 80

# Define environment variable
ENV NAME npm

# Run the npm command by default when the container starts.
ENTRYPOINT ./npm