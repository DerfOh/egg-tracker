import { render, screen, fireEvent } from "@testing-library/react";
import OptionsForm from "../src/main.jsx";

describe("OptionsForm validation", () => {
  it("shows error if name is empty", async () => {
    render(<OptionsForm type="species" option={null} onClose={() => {}} onSaved={() => {}} />);
    fireEvent.change(screen.getByLabelText(/name/i), { target: { value: "" } });
    fireEvent.click(screen.getByText(/save/i));
    expect(await screen.findByText(/name is required/i)).toBeInTheDocument();
  });

  it("calls onSaved on valid submit", async () => {
    const onSaved = jest.fn();
    global.fetch = jest.fn(() => Promise.resolve({ ok: true, json: () => Promise.resolve({}) }));
    render(<OptionsForm type="species" option={null} onClose={() => {}} onSaved={onSaved} />);
    fireEvent.change(screen.getByLabelText(/name/i), { target: { value: "Duck" } });
    fireEvent.click(screen.getByText(/save/i));
    // Wait for async
    await new Promise(r => setTimeout(r, 10));
    expect(onSaved).toHaveBeenCalled();
  });
});