import React from "react";
import { render, fireEvent, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

// Minimal AuthContext mock
const AuthContext = React.createContext({ login: jest.fn(), loading: false, error: null });
function AuthProvider({ children, login = jest.fn() }) {
  return (
    <AuthContext.Provider value={{ login, loading: false, error: null }}>
      {children}
    </AuthContext.Provider>
  );
}
function useAuth() {
  return React.useContext(AuthContext);
}

// Inline LoginForm from main.jsx
function LoginForm() {
  const { login, loading, error } = useAuth();
  const [email, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [formError, setFormError] = React.useState(null);
  const handleSubmit = async (e) => {
    e.preventDefault();
    setFormError(null);
    if (!email || !password) {
      setFormError("Email and password required");
      return;
    }
    await login(email, password);
  };
  return (
    <form onSubmit={handleSubmit}>
      <input
        type="email"
        placeholder="Email"
        value={email}
        onChange={e => setEmail(e.target.value)}
        required
      />
      <input
        type="password"
        placeholder="Password"
        value={password}
        onChange={e => setPassword(e.target.value)}
        required
      />
      {(formError || error) && <div>{formError || error}</div>}
      <button type="submit" disabled={loading}>Login</button>
    </form>
  );
}

describe("LoginForm", () => {
  it("shows error if fields are empty", async () => {
    render(
      <AuthProvider>
        <LoginForm />
      </AuthProvider>
    );
    fireEvent.click(screen.getByText("Login"));
    expect(await screen.findByText(/required/)).toBeInTheDocument();
  });

  it("calls login with email and password", async () => {
    const login = jest.fn();
    render(
      <AuthProvider login={login}>
        <LoginForm />
      </AuthProvider>
    );
    fireEvent.change(screen.getByPlaceholderText("Email"), { target: { value: "a@b.com" } });
    fireEvent.change(screen.getByPlaceholderText("Password"), { target: { value: "secret123" } });
    fireEvent.click(screen.getByText("Login"));
    // login should be called
    expect(login).toHaveBeenCalledWith("a@b.com", "secret123");
  });
});
