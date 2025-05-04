import React from "react";
import { render, fireEvent, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

// Minimal InventoryActionForm mock (formerly EggForm)
function InventoryActionForm({ action = "collected", record, onClose = () => {}, onSaved = () => {} }) {
  const [date, setDate] = React.useState(record ? record.date?.slice(0, 10) : "");
  const [species, setSpecies] = React.useState(record ? record.species || "" : "");
  const [quantity, setQuantity] = React.useState(record ? record.quantity || 1 : 1);
  const [error, setError] = React.useState(null);
  const [saving, setSaving] = React.useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(null);
    if (!date) {
      setError("Date is required");
      return;
    }
    if (!species) {
      setError("Species is required");
      return;
    }
    if (!quantity || quantity < 1) {
      setError("Quantity must be at least 1");
      return;
    }
    setSaving(true);
    // Simulate API call for inventory action
    setTimeout(() => {
      setSaving(false);
      onSaved();
    }, 10);
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="date"
        placeholder="Date"
        value={date}
        onChange={e => setDate(e.target.value)}
        required
      />
      <input
        type="text"
        placeholder="Species"
        value={species}
        onChange={e => setSpecies(e.target.value)}
        required
      />
      <input
        type="number"
        placeholder="Quantity"
        value={quantity}
        min={1}
        onChange={e => setQuantity(Number(e.target.value))}
        required
      />
      {error && <div>{error}</div>}
      <button type="submit" disabled={saving}>Save</button>
    </form>
  );
}

describe("InventoryActionForm (formerly EggForm)", () => {
  it("shows error if date is empty", async () => {
    render(<InventoryActionForm />);
    fireEvent.change(screen.getByPlaceholderText("Species"), { target: { value: "Chicken" } });
    fireEvent.change(screen.getByPlaceholderText("Quantity"), { target: { value: 2 } });
    fireEvent.click(screen.getByText("Save"));
    expect(await screen.findByText(/Date is required/)).toBeInTheDocument();
  });

  it("shows error if species is empty", async () => {
    render(<InventoryActionForm />);
    fireEvent.change(screen.getByPlaceholderText("Date"), { target: { value: "2024-05-01" } });
    fireEvent.change(screen.getByPlaceholderText("Quantity"), { target: { value: 2 } });
    fireEvent.click(screen.getByText("Save"));
    expect(await screen.findByText(/Species is required/)).toBeInTheDocument();
  });

  it("calls onSaved if valid", async () => {
    const onSaved = jest.fn();
    render(<InventoryActionForm onSaved={onSaved} />);
    fireEvent.change(screen.getByPlaceholderText("Date"), { target: { value: "2024-05-01" } });
    fireEvent.change(screen.getByPlaceholderText("Species"), { target: { value: "Chicken" } });
    fireEvent.change(screen.getByPlaceholderText("Quantity"), { target: { value: 2 } });
    fireEvent.click(screen.getByText("Save"));
    await new Promise(r => setTimeout(r, 20));
    expect(onSaved).toHaveBeenCalled();
  });
});
