import json
import requests
import firebase_admin
from firebase_admin import credentials, auth

# Replace with your Firebase project credentials
api_key = "AIzaSyA8PMxaKKasISGvPO6az2jhHJ5Tgh5BRfI"
firebase_project_id = "roundtown-a0579"


def sign_in_with_email_and_password(email, password, return_secure_token=True):
    payload = json.dumps({"email":email, "password":password, "return_secure_token":return_secure_token})
    FIREBASE_WEB_API_KEY = 'AIzaSyA8PMxaKKasISGvPO6az2jhHJ5Tgh5BRfI' 
    rest_api_url = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword"

    r = requests.post(rest_api_url,
                  params={"key": FIREBASE_WEB_API_KEY},
                  data=payload)

    return r.json()

# Initialize Firebase app
cred = credentials.Certificate("service_account.json")
firebase_admin.initialize_app(cred, {
    'apiKey': api_key,
    'projectId': firebase_project_id,
})

# Get user choice
choice = input("Do you want to sign in (1) or sign up (2)? ")

if choice == "1":
    # Sign in
    email = input("Enter your email: ")
    password = input("Enter your password: ")

    try:
        user = sign_in_with_email_and_password(email, password)
        id_token = user['idToken']
        print("ID token:", id_token)
    except Exception as e:
        print("Error:", e)

elif choice == "2":
    # Sign up
    email = input("Enter your email: ")
    password = input("Enter your password: ")
    display_name = input("Enter username: ")

    try:
        user = auth.create_user(email=email, password=password, display_name=display_name)
        print("User created successfully!")
    except Exception as e:
        print("Error:", e)

else:
    print("Invalid choice. Please enter 1 or 2.")