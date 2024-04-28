use std::sync::{Arc, Mutex};
use eframe::egui;

mod api;

fn main() -> Result<(), eframe::Error> {
    let api = api::new("http://localhost:3000".to_string());

    let options = eframe::NativeOptions {
        viewport: egui::ViewportBuilder::default().with_inner_size([320.0, 240.0]),
        ..Default::default()
    };

    // Our application state:
    let mut username: String = "replace@gmail.com".to_owned();
    let mut password: String = "replace123".to_owned();

    let token: Arc<Mutex<String>> = Arc::new(Mutex::new("".to_string()));

    eframe::run_simple_native("NextNet GUI", options, move |ctx, _frame| {
        egui::CentralPanel::default().show(ctx, |ui| {
            ui.heading("Login");
            ui.horizontal(|ui| {
                ui.label("Email: ");
                ui.text_edit_singleline(&mut username);
            });
            
            ui.horizontal(|ui| {
                let label = ui.label("Password: ");
                ui.add(egui::TextEdit::singleline(&mut password).password(true))
                    .labelled_by(label.id);
            });

            if ui.button("Login").clicked() {
                let token_clone = Arc::clone(&token);
                api.login(username.as_str(), password.as_str(), Box::new(move |res: api::LoginResponse| {
                    match res.token {
                        Some(x) => {
                            let mut token = token_clone.lock().unwrap();
                            *token = x;
                        },
                        None => {
                            let mut token = token_clone.lock().unwrap();
                            *token = "".to_string();
                        }
                    }
                }));
            }

            match token.lock() {
                Ok(x) => {
                    ui.label(format!("Token: {:?}", *x));
                },
                Err(_) => {
                    ui.label(format!("No token."));
                }
            }
        });
    })
}