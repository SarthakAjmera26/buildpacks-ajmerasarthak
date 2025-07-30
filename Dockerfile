# Use a more general-purpose base image like Ubuntu.
FROM ubuntu:latest

# Switch to root user for global installations (Ubuntu defaults to root)
# USER root # This line is often redundant in Ubuntu images but harmless.

# Install necessary dependencies: Python3, pip, git, and software-properties-common.
# We'll install Node.js and npm from a specific PPA later.
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    git \
    curl \
    software-properties-common \
    # Clean up apt caches to keep the image size down
    && rm -rf /var/lib/apt/lists/*

# Install Node.js v20 (or newer) using NodeSource PPA to satisfy @google/gemini-cli requirements
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs && \
    rm -rf /var/lib/apt/lists/* # Clean up apt lists again after Node.js installation

# The 'nodejs' package from NodeSource PPA usually creates the /usr/bin/node symlink correctly,
# so we no longer need to explicitly create it here.
# RUN ln -s /usr/bin/nodejs /usr/bin/node

# Install the Cloud Native Buildpacks 'pack' CLI using its official PPA.
RUN add-apt-repository ppa:cncf-buildpacks/pack-cli -y && \
    apt-get update && \
    apt-get install -y pack-cli && \
    rm -rf /var/lib/apt/lists/* # Clean up apt lists again after pack-cli installation

# Install Docker inside the container
RUN install -m 0755 -d /etc/apt/keyrings && \
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg && \
    chmod a+r /etc/apt/keyrings/docker.gpg && \
    echo \
    "deb [arch=\"$(dpkg --print-architecture)\" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
    \"$(. /etc/os-release && echo \"$VERSION_CODENAME\")\" stable" | \
    tee /etc/apt/sources.list.d/docker.list > /dev/null && \
    apt-get update && \
    apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin && \
    rm -rf /var/lib/apt/lists/*

# Set the working directory inside the container
WORKDIR /app

# Copy the requirements file and install Python dependencies
COPY requirements.txt .
RUN pip install -r requirements.txt --break-system-packages # Use || true to prevent build failure if requirements.txt is empty or missing

# Install the Google Gemini CLI globally using npm
RUN npm install -g @google/gemini-cli

# Copy the rest of your application code into the container
COPY . .

# Configure Git globally for safe directories and user identity
RUN git config --global --add safe.directory /app
RUN git config --global user.email "ajmerasarthak@gmail.com"
RUN git config --global user.name "SarthakAjmera26"

# Define the command to run when the container starts
# CMD ["python3", "run_gemini.py", "Try to integrate changes from https://rubyreferences.github.io/rubychanges/3.4.html#default-gems-that-became-bundled into the repo. Validate yourself by running `python3 validate.py --app-dir ruby/generic/simple --version 3.4` and if needed correct yourself."]
# ENTRYPOINT [ "bash" ]