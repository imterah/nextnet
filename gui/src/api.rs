use ehttp::{fetch, Request};
use serde::Deserialize;
use serde_json::json;

pub struct NextAPIClient {
    pub url: String,
}

#[derive(Deserialize, Debug, Default)]
pub struct LoginResponse {
    pub error: Option<String>,
    pub token: Option<String>,
}

impl NextAPIClient {
    pub fn login(
        &self,
        email: &str,
        password: &str,
        mut callback: impl 'static + Send + FnMut(LoginResponse),
    ) {
        let json_data = json!({
            "email": email,
            "password": password
        });

        let json_str_raw: String = json_data.to_string();
        let json_str: &str = json_str_raw.as_str();

        println!("{}", json_str);

        let mut request = Request::post(
            self.url.clone() + "/api/v1/users/login",
            json_str.as_bytes().to_vec(),
        );
        request.headers.insert("Content-Type", "application/json");

        fetch(request, move |result: ehttp::Result<ehttp::Response>| {
            let res = result.unwrap();
            let json: LoginResponse = res.json().unwrap();
            callback(json);
        });
    }
}

pub fn new(url: String) -> NextAPIClient {
    let api_client: NextAPIClient = NextAPIClient { url };

    return api_client;
}
