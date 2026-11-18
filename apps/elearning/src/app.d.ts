declare global {
  namespace App {
    interface PageState {
      // Pesan kegagalan /sso yang ditampilkan di /login tanpa menaruhnya di URL.
      ssoError?: string;
    }
  }
}
export {};
