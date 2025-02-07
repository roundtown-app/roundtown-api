import requests
import json
import os
from collections import defaultdict
import firebase_admin
from firebase_admin import credentials, auth
import psycopg2
from psycopg2 import sql

# Firebase configuration
FIREBASE_WEB_API_KEY = 'AIzaSyA8PMxaKKasISGvPO6az2jhHJ5Tgh5BRfI'
FIREBASE_PROJECT_ID = 'roundtown-a0579'

# PostgreSQL configuration
DB_NAME = 'postgres'
DB_USER = 'postgres'
DB_PASSWORD = 'password1'
DB_HOST = 'localhost'
DB_PORT = '5432'

# Function to sign in user and get token
def sign_in_with_email_and_password(email, password):
    payload = json.dumps({"email": email, "password": password, "return_secure_token": True})
    rest_api_url = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword"
    r = requests.post(rest_api_url,
                      params={"key": FIREBASE_WEB_API_KEY},
                      data=payload)
    return r.json()

# Initialize Firebase app
cred = credentials.Certificate("service_account.json")
firebase_admin.initialize_app(cred, {
    'projectId': FIREBASE_PROJECT_ID,
})

# Function to connect to PostgreSQL
def get_db_connection():
    return psycopg2.connect(
        dbname=DB_NAME,
        user=DB_USER,
        password=DB_PASSWORD,
        host=DB_HOST,
        port=DB_PORT
    )

# Function to update account type in PostgreSQL
def update_account_type(user_id, account_type):
    conn = get_db_connection()
    cur = conn.cursor()
    try:
        cur.execute(
            sql.SQL("UPDATE users SET account_type = %s WHERE firebase_uid = %s"),
            (account_type, user_id)
        )
        conn.commit()
        print(f"Updated account type to '{account_type}' for user {user_id}")
    except Exception as e:
        print(f"Error updating account type: {e}")
    finally:
        cur.close()
        conn.close()

# Function to get bearer token and update account type if necessary
def get_bearer_token(email, password, account_type):
    try:
        user = sign_in_with_email_and_password(email, password)
        id_token = user['idToken']
        print(f"Successfully signed in as {account_type}")
        
        # Update account type for business and admin accounts
        if account_type in ['business', 'admin']:
            firebase_user = auth.get_user_by_email(email)
            update_account_type(firebase_user.uid, account_type)
        
        return id_token
    except Exception as e:
        print(f"Error signing in as {account_type}: {e}")
        return None

# Get bearer tokens
USER_BEARER_TOKEN = get_bearer_token("testuser2@test.com", "password", "user")
BUSINESS_BEARER_TOKEN = get_bearer_token("testbusiness2@test.com", "password", "business")
ADMIN_BEARER_TOKEN = get_bearer_token("testadmin2@test.com", "password", "admin")

# Base URL for the API
BASE_URL = "http://localhost:8080"

# Dictionary to store created IDs
created_ids = defaultdict(dict)

def make_request(method, endpoint, token, json_body=None):
    headers = {"Authorization": f"Bearer {token}"}
    url = f"{BASE_URL}{endpoint}"
    
    if method == "GET":
        response = requests.get(url, headers=headers)
    elif method == "POST":
        response = requests.post(url, headers=headers, json=json_body)
    elif method == "PUT":
        response = requests.put(url, headers=headers, json=json_body)
    elif method == "DELETE":
        response = requests.delete(url, headers=headers)
    else:
        raise ValueError(f"Unsupported HTTP method: {method}")
    
    return response

def save_response(response, endpoint, method, token):
    if len(response.content):
        endpoint_file = str(endpoint).replace("{", "").replace("}", "")
        if endpoint_file[-1] != "/":
            endpoint_file += "/"
        else:
            endpoint_file += "-"
        response_file = f"./test-response-json{endpoint_file}{method.lower()}-{token}.json"
        os.makedirs(os.path.dirname(response_file), exist_ok=True)
        with open(response_file, 'w') as f:
            json.dump(response.json(), f, indent=2)
        print(f"  Response saved to: {response_file}")
    else:
        print('  No response body')

def test_endpoint(method, endpoint, tokens, resource_type=None):
    print(f"Testing {method} {endpoint}")
    
    # Determine if there's a JSON body for this request
    endpoint_file = str(endpoint).replace("{", "").replace("}", "")
    if endpoint_file[-1] != "/":
        endpoint_file += "/"
    else:
        endpoint_file += "-"
    json_file = f"./test-request-json{endpoint_file}{method.lower()}.json"
    json_body = None
    if os.path.exists(json_file):
        with open(json_file, 'r') as f:
            json_body = json.load(f)
    else:
        print(json_file + " not found")
    
    for token in tokens:
        token_acct = ""
        if token == USER_BEARER_TOKEN:
            token_acct = "user"
        elif token == BUSINESS_BEARER_TOKEN:
            token_acct = "business"
        elif token == ADMIN_BEARER_TOKEN:
            token_acct = "admin"

        print(f"  Using token: {token_acct}")
        
        # Replace placeholders in the endpoint with actual IDs
        final_endpoint = endpoint
        if resource_type and '{' in endpoint:
            id_key = f"{resource_type}ID"
            if id_key in created_ids[token]:
                final_endpoint = endpoint.replace(f"{{{id_key}}}", created_ids[token][id_key])
            else:
                print(f"  Skipping: No {id_key} available for this token")
                continue
        
        response = make_request(method, final_endpoint, token, json_body)
        
        print(f"  Status code: {response.status_code}")

        if response.status_code < 299:
            save_response(response, final_endpoint, method, token_acct)
        
        # If this is a POST request, store the created ID
        if method == "POST" and response.status_code == 201 and resource_type:
            created_id = response.json().get('id')
            if created_id:
                created_ids[token][f"{resource_type}ID"] = str(created_id)
                print(f"  Stored {resource_type}ID: {created_id}")

        # If this is the api/users/self endpoint, store the created ID
        if final_endpoint == "api/users/self" and response.status_code == 200:
            created_id = response.json().get('UserId')
            if created_id:
                created_ids[token][f"{resource_type}ID"] = str(created_id)
                print(f"  Stored {resource_type}ID: {created_id}")
        
        print()

# List of all endpoints to test, grouped by resource type
endpoints = [
    # User endpoints
    # ("POST", "/api/users", "user"),
    ("GET", "/api/users/self", "user"),
    ("GET", "/api/users/{userID}", "user"),
    # ("PUT", "/api/users", "user"),
    ("PUT", "/api/users/updateLocation", "user"),
    ("POST", "/api/users/search", "user"),

    # Venue endpoints
    ("POST", "/api/venues", "venue"),
    ("GET", "/api/venues/{venueID}", "venue"),
    ("PUT", "/api/venues/{venueID}", "venue"),
    ("POST", "/api/venues/search", "venue"),
    ("DELETE", "/api/venues/{venueID}", "venue"),
    
    # Event endpoints
    ("POST", "/api/events", "event"),
    ("GET", "/api/events/{eventID}", "event"),
    ("PUT", "/api/events/{eventID}", "event"),
    ("POST", "/api/events/search", "event"),
    ("DELETE", "/api/events/{eventID}", "event"),
    
    # Plan endpoints
    ("POST", "/api/plans", "plan"),
    ("GET", "/api/plans/{planID}", "plan"),
    ("PUT", "/api/plans/{planID}", "plan"),
    ("POST", "/api/plans/search", "plan"),
    ("DELETE", "/api/plans/{planID}", "plan"),
    
    # Recommendation endpoints
    ("GET", "/api/recommendations/feed-recs/{userID}", "user"),
    ("GET", "/api/recommendations/plan-recs/{userID}", "user"),
    ("GET", "/api/recommendations/event-based-recs/{userID}", "user"),
    
    # Interaction endpoints
    ("PUT", "/api/interactions/saveItem", None),
    ("GET", "/api/interactions/savedItems", None),
    ("POST", "/api/interactions/unsaveItem", None),
    ("PUT", "/api/interactions/subscribeItem", None),
    ("GET", "/api/interactions/subscribedItems", None),
    ("POST", "/api/interactions/unsubscribeItem", None),
    ("PUT", "/api/interactions/visitItem", None),
    ("GET", "/api/interactions/visitedItems", None),
    ("PUT", "/api/interactions/shareItem", None),
    ("GET", "/api/interactions/sharedItems", None),

    # User delete endpoint
    ("DELETE", "/api/users", "user"),
]

# Test all endpoints
for method, endpoint, resource_type in endpoints:
    test_endpoint(method, endpoint, [USER_BEARER_TOKEN, BUSINESS_BEARER_TOKEN, ADMIN_BEARER_TOKEN], resource_type)

print("All tests completed.")