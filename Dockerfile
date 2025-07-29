# Use an official Python runtime as a parent image
FROM nikolaik/python-nodejs:latest

# Switch to root user for global installations
USER root

# Set the working directory to /app
# This is a standard and recommended practice for applications.
WORKDIR /app

# Install any Python dependencies specified in requirements.txt
# Now running as root, so it should have permissions to install globally
COPY requirements.txt .
RUN pip install -r requirements.txt || true # Use || true to prevent build failure if requirements.txt is empty or missing

# Install the Gemini CLI globally using npm
# Now running as root, so it should have permissions
RUN npm install -g @google/gemini-cli

# Copy the current directory contents into the container at the current WORKDIR (/app)
# Files will be copied into '/app' and owned by 'pn'.
COPY . .
RUN git config --global --add safe.directory /app
RUN git config --global user.email "ajmerasarthak@gmail.com"
RUN git config --global user.name "SarthakAjmera26"
# Command to run the application with the specified arguments
CMD ["python3", "run_gemini.py", "create a new file named hello-00.txt in buildpacks/cmd/go", "--repo", "https://github.com/SarthakAjmera26/buildpacks-ajmerasarthak", "--branch", "main", "--title", "Gemini: Create hello-00.txt"]
