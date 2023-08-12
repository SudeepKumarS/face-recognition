from fastapi import FastAPI, UploadFile, File
from fastapi.responses import JSONResponse
import uvicorn
import face_recognition
import logging
import traceback

app = FastAPI()

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Create a formatter
formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')

# Creating a file handler to store the logs
fh = logging.FileHandler('app.log')
# setting the level to INFO
fh.setLevel(logging.INFO)
fh.setFormatter(formatter)

# Adding the handler to the logger
logger.addHandler(fh)


# Root Homepage
@app.get("/")
def root_page():
    return {
        "Hi!": "This is a dev face recogntion server!"
    }

# Endpoint to compare images
@app.post("/face-recognition")
async def compare_faces(file1: UploadFile = File(...), file2: UploadFile = File(...)):
    try:
        logger.info("Received a request to compare faces.")

        # Reading the uploaded images
        image1 = face_recognition.load_image_file(file1.file)
        image2 = face_recognition.load_image_file(file2.file)

        # Generating face encodings for the uploaded images
        encoding1 = face_recognition.face_encodings(image1)[0]
        encoding2 = face_recognition.face_encodings(image2)[0]

        # Comparing the face encodings
        distances = face_recognition.compare_faces([encoding1], encoding2)[0]

        # Using bool datatype to turn the numpy.bool_
        result = {
            "matched": bool(distances)
        }

        logger.info("Face comparison successful.")
        # Output the distance in JSON format
        return JSONResponse(status_code=200, content=result)
    except Exception as e:
        error_message = "An error occurred while processing the request."
        logger.error(f"{error_message}\n{traceback.format_exc()}")
        return JSONResponse(status_code=400, content={"error": error_message})


if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=8000)