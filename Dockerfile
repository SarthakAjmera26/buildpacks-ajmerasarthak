# Use an official Python runtime as a parent image
FROM nikolaik/python-nodejs:latest

# Switch to root user for global installations
USER root

WORKDIR /app

COPY requirements.txt .
RUN pip install -r requirements.txt || true # Use || true to prevent build failure if requirements.txt is empty or missing

RUN npm install -g @google/gemini-cli

COPY . .
RUN git config --global --add safe.directory /app
RUN git config --global user.email "ajmerasarthak@gmail.com"
RUN git config --global user.name "SarthakAjmera26"

