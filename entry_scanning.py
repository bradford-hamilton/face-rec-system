import cv2
import requests
import time
import signal
from threading import Timer

class Color:
    RED = '\033[31m'
    GREEN = '\033[32m'
    BLUE = '\033[34m'
    RESET = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'

def pretty_print(text, color=Color.RESET, bold=False):
    style = color + (Color.BOLD if bold else '')
    print(style + text + color)

def signal_handler(signum, frame):
    global keep_running
    print("Termination signal received, shutting down...")
    keep_running = False
    cap.release()

def capture_and_send():
    if not keep_running:
        return

    # Capture image (read single frame)
    ret, frame = cap.read()
    if not ret:
        print("Failed to capture image")
        return

    # Save the captured frame as JPEG
    _, buffer = cv2.imencode('.jpeg', frame)

    # Convert to bytes and send as multipart/form-data
    files = {'image': ('image.jpeg', buffer.tobytes(), 'image/jpeg')}
    response = requests.post(match_handler_url, files=files)

    if response.status_code == 200:
        data = response.json()
        pretty_print(f"Match found! User ID: {data['user_id']}, Email: {data['email']}", Color.GREEN, True)
    elif response.status_code == 404:
        pretty_print("No matches found")
    else:
        print("An error occurred", Color.RED, False)

    # Schedule the next capture if we are still running
    if keep_running:
        Timer(capture_interval, capture_and_send).start()

# ------------------------ main ------------------------

capture_interval = 5 # seconds
match_handler_url = "http://localhost:4000/match"

# Initialize the camera
cap = cv2.VideoCapture(0)

# Flag to control the capture loop
keep_running = True

# Register signal handlers
signal.signal(signal.SIGINT, signal_handler)
signal.signal(signal.SIGTERM, signal_handler)

# Start the periodic image capture and send
capture_and_send()

# Keep the script running
while keep_running:
    time.sleep(1)
