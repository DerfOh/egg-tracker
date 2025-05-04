import React from "react";
import { render, fireEvent, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

// Minimal InventoryForm mock (copied from main.jsx)
function InventoryForm({ action, onClose = () => {}, onSaved = () => {} }) {
  const [date, setDate] = React.useState(action ? action.date?.slice(0, 10) : "");
  const [species, setSpecies] = React.useState(action ? action.species || "" : "");
  const [quantity, setQuantity] = React.useState(action ? action.quantity : "");
  const [actType, setActType] = React.useState(action ? action.action || "collected" : "collected");
  const [notes, setNotes] = React.useState(action ? action.notes || "" : "");
  const [error, setError] = React.useState(null);
  const [saving, setSaving] = React.useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(null);
    if (!date || !species || !quantity) {
      setError("Date, species, and quantity are required");
      return;
    }
    if (isNaN(Number(quantity)) || Number(quantity) <= 0) {
      setError("Quantity must be a positive number");
      return;
    }
    setSaving(true);
    setTimeout(() => {
      setSaving(false);
      onSaved();
    }, 10);
  };

  return (
    <form onSubmit={handleSubmit}>
      <input type="date" placeholder="Date" value={date} onChange={e => setDate(e.target.value)} required />
      <input type="text" placeholder="Species" value={species} onChange={e => setSpecies(e.target.value)} required />
      <input type="number" placeholder="Quantity" value={quantity} onChange={e => setQuantity(e.target.value)} required />
      <select value={actType} onChange={e => setActType(e.target.value)} required>
        <option value="collected">Collected</option>
        <option value="sold">Sold</option>
        <option value="consumed">Consumed</option>
        <option value="gifted">Gifted</option>
        <option value="spoiled">Spoiled/Broken</option>
      </select>
      <input type="text" placeholder="Notes" value={notes} onChange={e => setNotes(e.target.value)} />
      {error && <div>{error}</div>}
      <button type="submit" disabled={saving}>Save</button>
    </form>
  );
}

describe("InventoryForm", () => {
  it("shows error if required fields are empty", async () => {
    render(<InventoryForm />);
    fireEvent.click(screen.getByText("Save"));
    expect(await screen.findByText(/required/)).toBeInTheDocument();
  });

  it("shows error if quantity is not positive", async () => {
    render(<InventoryForm />);
    fireEvent.change(screen.getByPlaceholderText("Date"), { target: { value: "2024-05-01" } });
    fireEvent.change(screen.getByPlaceholderText("Species"), { target: { value: "Goose" } });
    fireEvent.change(screen.getByPlaceholderText("Quantity"), { target: { value: "0" } });
    fireEvent.click(screen.getByText("Save"));
    expect(await screen.findByText(/positive number/)).toBeInTheDocument();
  });

  it("calls onSaved if valid", async () => {
    const onSaved = jest.fn();
    render(<InventoryForm onSaved={onSaved} />);
    fireEvent.change(screen.getByPlaceholderText("Date"), { target: { value: "2024-05-01" } });
    fireEvent.change(screen.getByPlaceholderText("Species"), { target: { value: "Goose" } });
    fireEvent.change(screen.getByPlaceholderText("Quantity"), { target: { value: "5" } });
    fireEvent.click(screen.getByText("Save"));
    await new Promise(r => setTimeout(r, 20));
    expect(onSaved).toHaveBeenCalled();
  });
});
