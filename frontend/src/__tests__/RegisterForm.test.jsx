import React from "react";
import { render, fireEvent, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

// Minimal AuthContext mock
const AuthContext = React.createContext({ register: jest.fn(), loading: false, error: null });
function AuthProvider({ children, register = jest.fn() }) {
  return (
    <AuthContext.Provider value={{ register, loading: false, error: null }}>
      {children}
    </AuthContext.Provider>
  );
}
function useAuth() {
  return React.useContext(AuthContext);
}

// Inline RegisterForm from main.jsx
function RegisterForm() {
  const { register, loading, error } = useAuth();
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
    if (password.length < 8) {
      setFormError("Password must be at least 8 characters");
      return;
    }
    await register(email, password);
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
        placeholder="Password (min 8 chars)"
        value={password}
        onChange={e => setPassword(e.target.value)}
        required
      />
      {(formError || error) && <div>{formError || error}</div>}
      <button type="submit" disabled={loading}>Register</button>
    </form>
  );
}

describe("RegisterForm", () => {
  it("shows error if fields are empty", async () => {
    render(
      <AuthProvider>
        <RegisterForm />
      </AuthProvider>
    );
    fireEvent.click(screen.getByText("Register"));
    expect(await screen.findByText(/required/)).toBeInTheDocument();
  });

  it("shows error if password is too short", async () => {
    render(
      <AuthProvider>
        <RegisterForm />
      </AuthProvider>
    );
    fireEvent.change(screen.getByPlaceholderText("Email"), { target: { value: "a@b.com" } });
    fireEvent.change(screen.getByPlaceholderText("Password (min 8 chars)"), { target: { value: "short" } });
    fireEvent.click(screen.getByText("Register"));
    expect(await screen.findByText(/at least 8/)).toBeInTheDocument();
  });

  it("calls register with email and password", async () => {
    const register = jest.fn();
    render(
      <AuthProvider register={register}>
        <RegisterForm />
      </AuthProvider>
    );
    fireEvent.change(screen.getByPlaceholderText("Email"), { target: { value: "a@b.com" } });
    fireEvent.change(screen.getByPlaceholderText("Password (min 8 chars)"), { target: { value: "longenough" } });
    fireEvent.click(screen.getByText("Register"));
    expect(register).toHaveBeenCalledWith("a@b.com", "longenough");
  });
});
