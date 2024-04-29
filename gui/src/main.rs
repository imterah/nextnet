use std::sync::{Arc, Mutex};
use eframe::egui;

mod components;
mod api;

pub struct ApplicationState {
    token: Arc<Mutex<String>>,
    username: String,
    password: String,
}

fn main() -> Result<(), eframe::Error> {
    let api = api::new("http://localhost:3000".to_string());

    let options = eframe::NativeOptions {
        viewport: egui::ViewportBuilder::default().with_inner_size([1280.0, 720.0]),
        ..Default::default()
    };

    let mut app_state: ApplicationState = ApplicationState {
        token: Arc::new(Mutex::new("".to_string())),
        
        // /!\ NOT THREAD SAFE FIELDS /!\
        // These are used internally for each application (immediate mode + functions which are stateless,
        // and we need *a* state somehow) 

        // components/log_in.rs
        username: "replace@gmail.com".to_owned(),
        password: "replace123".to_owned()
    };

    eframe::run_simple_native("NextNet GUI", options, move |ctx, _frame| {
        egui::CentralPanel::default().show(ctx, |_ui| {
            components::log_in::main(&mut app_state, &api, ctx);
        });
    })
}