import React, { useEffect, useState, createContext, useContext } from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter as Router, Routes, Route, Link, useNavigate, Navigate, Outlet } from "react-router-dom";
import "./index.css"; // TailwindCSS should be imported here

// Set API base URL
const BASE_API = ""; // Keep this as relative path for API calls

// --- AuthContext for login/register, token in memory, refresh logic ---
const AuthContext = createContext();

function AuthProvider({ children }) {
  const [accessToken, setAccessToken] = useState(null);
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Try refresh on mount if no token
  useEffect(() => {
    if (!accessToken) {
      fetch(BASE_API + "/api/refresh", { method: "POST", credentials: "include" })
        .then(async (res) => {
          if (res.ok) {
            const data = await res.json();
            setAccessToken(data.access_token);
          }
        })
        .catch(() => {});
    }
  }, []);

  const login = async (email, password) => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(BASE_API + "/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ email, password }),
      });
      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || "Login failed");
      }
      const data = await res.json();
      setAccessToken(data.access_token);
      setUser({ email });
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  const register = async (email, password) => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(BASE_API + "/api/signup", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || "Registration failed");
      }
      // Optionally auto-login after register
      await login(email, password);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  const logout = () => {
    setAccessToken(null);
    setUser(null);
    // Optionally: clear refresh cookie by calling a logout endpoint
  };

  return (
    <AuthContext.Provider value={{ accessToken, user, login, register, logout, loading, error }}>
      {children}
    </AuthContext.Provider>
  );
}

function useAuth() {
  return useContext(AuthContext);
}

// --- ThemeContext (unchanged) ---
const ThemeContext = createContext();

function ThemeProvider({ children }) {
  const [theme, setTheme] = useState(() => {
    const stored = localStorage.getItem("theme");
    if (stored) return stored;
    return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  });

  useEffect(() => {
    document.documentElement.classList.toggle("dark", theme === "dark");
    localStorage.setItem("theme", theme);
  }, [theme]);

  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

function ThemeToggle() {
  const { theme, setTheme } = useContext(ThemeContext);
  return (
    <button
      className="px-4 py-2 rounded bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100 mt-8"
      onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
    >
      Switch to {theme === "dark" ? "Light" : "Dark"} Mode
    </button>
  );
}

// --- Navbar component ---
function Navbar() {
  const { logout } = useAuth();
  const { theme, setTheme } = useContext(ThemeContext);
  const navigate = useNavigate();
  const handleBackup = async () => {
    await fetch(BASE_API + "/api/backup", { method: "POST", credentials: "include" });
    // Optionally show toast/alert
    alert("Backup triggered");
  };
  return (
    <nav className="flex items-center justify-between px-4 py-2 bg-gray-100 dark:bg-gray-800">
      <div className="flex gap-4">
        <Link to="/inventory" className="hover:underline">Inventory</Link>
        <Link to="/options" className="hover:underline">Options</Link>
        <Link to="/reports" className="hover:underline">Reports</Link>
      </div>
      <div className="flex gap-2 items-center">
        <button onClick={handleBackup} className="px-2 py-1 bg-yellow-500 text-white rounded">Backup</button>
        <button
          className="px-2 py-1 bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded"
          onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
        >
          {theme === "dark" ? "☀️" : "🌙"}
        </button>
        <button onClick={logout} className="px-2 py-1 bg-red-600 text-white rounded">Logout</button>
      </div>
    </nav>
  );
}

// --- InventoryPage (now main page, with species dropdown) ---
function InventoryPage() {
  const [actions, setActions] = useState([]); // Initialize as empty array
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [editAction, setEditAction] = useState(null);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [speciesOptions, setSpeciesOptions] = useState([]);
  const [coopOptions, setCoopOptions] = useState([]);
  const [colorOptions, setColorOptions] = useState([]);
  const [sizeOptions, setSizeOptions] = useState([]);

  useEffect(() => {
    setLoading(true);
    fetch(BASE_API + "/api/inventory", { credentials: "include" })
      .then(async (res) => {
        if (!res.ok) throw new Error("Failed to fetch inventory");
        return res.json();
      })
      .then(data => {
        if (data === null) {
          setActions([
            {
              id: 1,
              date: new Date().toISOString().slice(0, 10),
              species: "Goose",
              action: "collected",
              quantity: 5,
              notes: "Sample inventory action",
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
            },
          ]);
        } else {
          setActions(data);
        }
      })
      .catch(e => {
        setError(e.message);
        setActions([
          {
            id: 1,
            date: new Date().toISOString().slice(0, 10),
            species: "Goose",
            action: "collected",
            quantity: 5,
            notes: "Sample inventory action",
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          },
        ]);
      })
      .finally(() => setLoading(false));
  }, [showForm]);

  // Fetch options for dropdowns
  useEffect(() => {
    fetch(BASE_API + "/api/options/species", { credentials: "include" })
      .then(async (res) => {
        if (!res.ok) throw new Error("Failed to fetch species");
        return res.json();
      })
      .then(data => {
        setSpeciesOptions(Array.isArray(data) ? data.filter(s => s.active) : []);
      })
      .catch(() => {
        setSpeciesOptions([
          { id: 1, name: "Chicken" },
          { id: 2, name: "Goose" },
          { id: 3, name: "Guinea Fowl" },
        ]);
      });
    fetch(BASE_API + "/api/options/coop", { credentials: "include" })
      .then(async (res) => {
        if (!res.ok) throw new Error("Failed to fetch coops");
        return res.json();
      })
      .then(data => {
        setCoopOptions(Array.isArray(data) ? data.filter(s => s.active) : []);
      })
      .catch(() => {
        setCoopOptions([
          { id: 1, name: "Main Coop" },
          { id: 2, name: "Back Barn" },
        ]);
      });
    fetch(BASE_API + "/api/options/eggcolor", { credentials: "include" })
      .then(async (res) => {
        if (!res.ok) throw new Error("Failed to fetch egg colors");
        return res.json();
      })
      .then(data => {
        setColorOptions(Array.isArray(data) ? data.filter(s => s.active) : []);
      })
      .catch(() => {
        setColorOptions([
          { id: 1, name: "White" },
          { id: 2, name: "Brown" },
        ]);
      });
    fetch(BASE_API + "/api/options/eggsize", { credentials: "include" })
      .then(async (res) => {
        if (!res.ok) throw new Error("Failed to fetch egg sizes");
        return res.json();
      })
      .then(data => {
        setSizeOptions(Array.isArray(data) ? data.filter(s => s.active) : []);
      })
      .catch(() => {
        setSizeOptions([
          { id: 1, name: "Small" },
          { id: 2, name: "Large" },
        ]);
      });
  }, []);

  const pagedActions = actions.slice((page - 1) * pageSize, page * pageSize);
  const totalPages = Math.ceil(actions.length / pageSize);

  const handleEdit = (action) => {
    setEditAction(action);
    setShowForm(true);
  };
  const handleAdd = () => {
    setEditAction(null);
    setShowForm(true);
  };
  const handleDelete = async (action) => {
    if (!window.confirm("Delete this inventory action?")) return;
    await fetch(BASE_API + `/api/inventory/${action.id}`,
      { method: "DELETE", credentials: "include" });
    setActions(actions.filter(a => a.id !== action.id));
  };

  return (
    <div className="p-4">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-bold">Inventory</h2>
        <button onClick={handleAdd} className="bg-green-600 text-white px-3 py-1 rounded">Add Action</button>
      </div>
      {showForm && (
        <InventoryForm
          action={editAction}
          onClose={() => setShowForm(false)}
          onSaved={() => setShowForm(false)}
          speciesOptions={speciesOptions}
          coopOptions={coopOptions}
          colorOptions={colorOptions}
          sizeOptions={sizeOptions}
        />
      )}
      {loading ? (
        <div>Loading...</div>
      ) : error ? (
        <div className="text-red-500">{error}</div>
      ) : (
        <>
          <table className="min-w-full border mb-4">
            <thead>
              <tr className="bg-gray-200 dark:bg-gray-700">
                <th className="p-2 border">Date</th>
                <th className="p-2 border">Species</th>
                <th className="p-2 border">Action</th>
                <th className="p-2 border">Quantity</th>
                <th className="p-2 border">Notes</th>
                <th className="p-2 border">Actions</th>
              </tr>
            </thead>
            <tbody>
              {pagedActions.map(action => (
                <tr key={action.id} className="border-b">
                  <td className="p-2 border">{action.date?.slice(0, 10)}</td>
                  <td className="p-2 border">{action.species}</td>
                  <td className="p-2 border">{action.action}</td>
                  <td className="p-2 border">{action.quantity}</td>
                  <td className="p-2 border">{action.notes || "-"}</td>
                  <td className="p-2 border">
                    <button onClick={() => handleEdit(action)} className="px-2 py-1 bg-blue-500 text-white rounded mr-2">Edit</button>
                    <button onClick={() => handleDelete(action)} className="px-2 py-1 bg-red-600 text-white rounded">Delete</button>
                  </td>
                </tr>
              ))}
              {pagedActions.length === 0 && (
                <tr><td colSpan={6} className="p-2 text-center">No inventory actions found.</td></tr>
              )}
            </tbody>
          </table>
          <div className="flex gap-2 items-center">
            <button disabled={page === 1} onClick={() => setPage(p => p - 1)} className="px-2 py-1 border rounded">Prev</button>
            <span>Page {page} of {totalPages || 1}</span>
            <button disabled={page === totalPages || totalPages === 0} onClick={() => setPage(p => p + 1)} className="px-2 py-1 border rounded">Next</button>
          </div>
        </>
      )}
    </div>
  );
}

// InventoryForm with species dropdown
function InventoryForm({ action, onClose, onSaved, speciesOptions, coopOptions, colorOptions, sizeOptions }) {
  const [date, setDate] = useState(action ? action.date?.slice(0, 10) : "");
  const [species, setSpecies] = useState(action ? action.species || "" : "");
  const [coop, setCoop] = useState(action ? action.coop || "" : "");
  const [eggColor, setEggColor] = useState(action ? action.egg_color || "" : "");
  const [eggSize, setEggSize] = useState(action ? action.egg_size || "" : "");
  const [quantity, setQuantity] = useState(action ? action.quantity : "");
  const [actType, setActType] = useState(action ? action.action || "collected" : "collected");
  const [notes, setNotes] = useState(action ? action.notes || "" : "");
  const [error, setError] = useState(null);
  const [saving, setSaving] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(null);
    if (!date || !species || !quantity || !coop || !eggColor || !eggSize) {
      setError("Date, species, coop, egg color, egg size, and quantity are required");
      return;
    }
    if (isNaN(Number(quantity)) || Number(quantity) <= 0) {
      setError("Quantity must be a positive number");
      return;
    }
    setSaving(true);
    const payload = {
      date,
      species,
      coop,
      egg_color: eggColor,
      egg_size: eggSize,
      quantity: Number(quantity),
      action: actType,
      notes: notes || null,
    };
    try {
      let res;
      if (action) {
        res = await fetch(BASE_API + `/api/inventory/${action.id}`,
          {
            method: "PUT",
            headers: { "Content-Type": "application/json" },
            credentials: "include",
            body: JSON.stringify(payload),
          });
      } else {
        res = await fetch(BASE_API + "/api/inventory",
          {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            credentials: "include",
            body: JSON.stringify(payload),
          });
      }
      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || "Save failed");
      }
      onSaved();
    } catch (e) {
      setError(e.message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-30 flex items-center justify-center z-50">
      <form className="bg-white dark:bg-gray-800 p-6 rounded shadow flex flex-col gap-4 min-w-[300px]" onSubmit={handleSubmit}>
        <h3 className="text-lg font-bold mb-2">{action ? "Edit Inventory Action" : "Add Inventory Action"}</h3>
        <label>
          <span>Date *</span>
          <input type="date" className="p-2 border rounded w-full" value={date} onChange={e => setDate(e.target.value)} required />
        </label>
        <label>
          <span>Species *</span>
          <select className="p-2 border rounded w-full" value={species} onChange={e => setSpecies(e.target.value)} required>
            <option value="">Select species</option>
            {speciesOptions.map(opt => (
              <option key={opt.id} value={opt.name}>{opt.name}</option>
            ))}
          </select>
        </label>
        <label>
          <span>Coop *</span>
          <select className="p-2 border rounded w-full" value={coop} onChange={e => setCoop(e.target.value)} required>
            <option value="">Select coop</option>
            {coopOptions.map(opt => (
              <option key={opt.id} value={opt.name}>{opt.name}</option>
            ))}
          </select>
        </label>
        <label>
          <span>Egg Color *</span>
          <select className="p-2 border rounded w-full" value={eggColor} onChange={e => setEggColor(e.target.value)} required>
            <option value="">Select color</option>
            {colorOptions.map(opt => (
              <option key={opt.id} value={opt.name}>{opt.name}</option>
            ))}
          </select>
        </label>
        <label>
          <span>Egg Size *</span>
          <select className="p-2 border rounded w-full" value={eggSize} onChange={e => setEggSize(e.target.value)} required>
            <option value="">Select size</option>
            {sizeOptions.map(opt => (
              <option key={opt.id} value={opt.name}>{opt.name}</option>
            ))}
          </select>
        </label>
        <label>
          <span>Quantity *</span>
          <input type="number" min="1" className="p-2 border rounded w-full" value={quantity} onChange={e => setQuantity(e.target.value)} required />
        </label>
        <label>
          <span>Action *</span>
          <select className="p-2 border rounded w-full" value={actType} onChange={e => setActType(e.target.value)} required>
            <option value="collected">Collected</option>
            <option value="sold">Sold</option>
            <option value="consumed">Consumed</option>
            <option value="gifted">Gifted</option>
            <option value="spoiled">Spoiled/Broken</option>
          </select>
        </label>
        <label>
          <span>Notes</span>
          <input type="text" className="p-2 border rounded w-full" value={notes} onChange={e => setNotes(e.target.value)} />
        </label>
        {error && <div className="text-red-500">{error}</div>}
        <div className="flex gap-2 justify-end">
          <button type="button" onClick={onClose} className="px-3 py-1 bg-gray-300 rounded">Cancel</button>
          <button type="submit" className="px-3 py-1 bg-blue-600 text-white rounded" disabled={saving}>{saving ? "Saving..." : "Save"}</button>
        </div>
      </form>
    </div>
  );
}

function OptionsPage() {
  const [type, setType] = useState("species");
  const [options, setOptions] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [editOption, setEditOption] = useState(null);
  const [refresh, setRefresh] = useState(0);

  const typeLabels = {
    species: "Species",
    eggcolor: "Egg Colors",
    eggsize: "Egg Sizes",
    coop: "Coops/Barns",
  };

  // Sample data fallback
  const sampleOptions = {
    species: [
      { id: 1, name: "Chicken", active: true, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
      { id: 2, name: "Goose", active: true, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
    ],
    eggcolor: [
      { id: 1, name: "White", active: true, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
      { id: 2, name: "Brown", active: true, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
    ],
    eggsize: [
      { id: 1, name: "Small", active: true, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
      { id: 2, name: "Large", active: true, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
    ],
    coop: [
      { id: 1, name: "Main Coop", active: true, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
      { id: 2, name: "Back Barn", active: true, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
    ],
  };

  useEffect(() => {
    setLoading(true);
    setError(null);
    fetch(BASE_API + "/api/options/" + type, { credentials: "include" })
      .then(async (res) => {
        if (!res.ok) throw new Error("Failed to fetch options");
        return res.json();
      })
      .then((data) => {
        if (data === null) {
          setOptions(sampleOptions[type] || []);
        } else {
          setOptions(data);
        }
      })
      .catch((e) => {
        setError(e.message);
        setOptions(sampleOptions[type] || []);
      })
      .finally(() => setLoading(false));
  }, [type, showForm, refresh]);

  const handleAdd = () => {
    setEditOption(null);
    setShowForm(true);
  };
  const handleEdit = (opt) => {
    setEditOption(opt);
    setShowForm(true);
  };
  const handleDeactivate = async (opt) => {
    await fetch(BASE_API + `/api/options/${type}/${opt.id}/deactivate`, { method: "POST", credentials: "include" });
    setRefresh((r) => r + 1);
  };
  const handleReactivate = async (opt) => {
    await fetch(BASE_API + `/api/options/${type}/${opt.id}/reactivate`, { method: "POST", credentials: "include" });
    setRefresh((r) => r + 1);
  };

  return (
    <div className="p-4">
      <h2 className="text-xl font-bold mb-4">Options Management</h2>
      <div className="flex gap-2 mb-4">
        {Object.keys(typeLabels).map((t) => (
          <button
            key={t}
            onClick={() => setType(t)}
            className={`px-3 py-1 rounded ${type === t ? "bg-blue-600 text-white" : "bg-gray-200 dark:bg-gray-700"}`}
          >
            {typeLabels[t]}
          </button>
        ))}
      </div>
      <div className="flex justify-between items-center mb-2">
        <h3 className="text-lg font-semibold">{typeLabels[type]}</h3>
        <button onClick={handleAdd} className="bg-green-600 text-white px-3 py-1 rounded">Add</button>
      </div>
      {showForm && (
        <OptionsForm
          type={type}
          option={editOption}
          onClose={() => setShowForm(false)}
          onSaved={() => { setShowForm(false); setRefresh((r) => r + 1); }}
        />
      )}
      {loading ? (
        <div>Loading...</div>
      ) : error ? (
        <div className="text-red-500">{error}</div>
      ) : (
        <table className="min-w-full border mb-4">
          <thead>
            <tr className="bg-gray-200 dark:bg-gray-700">
              <th className="p-2 border">Name</th>
              <th className="p-2 border">Status</th>
              <th className="p-2 border">Actions</th>
            </tr>
          </thead>
          <tbody>
            {options.map((opt) => (
              <tr key={opt.id} className="border-b">
                <td className="p-2 border">{opt.name}</td>
                <td className="p-2 border">{opt.active ? "Active" : "Inactive"}</td>
                <td className="p-2 border">
                  <button onClick={() => handleEdit(opt)} className="px-2 py-1 bg-blue-500 text-white rounded mr-2">Edit</button>
                  {opt.active ? (
                    <button onClick={() => handleDeactivate(opt)} className="px-2 py-1 bg-yellow-600 text-white rounded">Deactivate</button>
                  ) : (
                    <button onClick={() => handleReactivate(opt)} className="px-2 py-1 bg-green-600 text-white rounded">Reactivate</button>
                  )}
                </td>
              </tr>
            ))}
            {options.length === 0 && (
              <tr><td colSpan={3} className="p-2 text-center">No options found.</td></tr>
            )}
          </tbody>
        </table>
      )}
    </div>
  );
}

function OptionsForm({ type, option, onClose, onSaved }) {
  const [name, setName] = useState(option ? option.name : "");
  const [error, setError] = useState(null);
  const [saving, setSaving] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(null);
    if (!name) {
      setError("Name is required");
      return;
    }
    setSaving(true);
    try {
      let res;
      if (option) {
        res = await fetch(BASE_API + `/api/options/${type}/${option.id}`,
          {
            method: "PUT",
            headers: { "Content-Type": "application/json" },
            credentials: "include",
            body: JSON.stringify({ name }),
          });
      } else {
        res = await fetch(BASE_API + `/api/options/${type}`,
          {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            credentials: "include",
            body: JSON.stringify({ name }),
          });
      }
      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || "Save failed");
      }
      onSaved();
    } catch (e) {
      setError(e.message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-30 flex items-center justify-center z-50">
      <form className="bg-white dark:bg-gray-800 p-6 rounded shadow flex flex-col gap-4 min-w-[300px]" onSubmit={handleSubmit}>
        <h3 className="text-lg font-bold mb-2">{option ? "Edit" : "Add"} {type.charAt(0).toUpperCase() + type.slice(1)}</h3>
        <label>
          <span>Name *</span>
          <input type="text" className="p-2 border rounded w-full" value={name} onChange={e => setName(e.target.value)} required />
        </label>
        {error && <div className="text-red-500">{error}</div>}
        <div className="flex gap-2 justify-end">
          <button type="button" onClick={onClose} className="px-3 py-1 bg-gray-300 rounded">Cancel</button>
          <button type="submit" className="px-3 py-1 bg-blue-600 text-white rounded" disabled={saving}>{saving ? "Saving..." : "Save"}</button>
        </div>
      </form>
    </div>
  );
}

function ReportsPage() {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [data, setData] = useState(null);
  const [speciesList, setSpeciesList] = useState([]);
  const [actionsList, setActionsList] = useState([
    "collected", "sold", "consumed", "gifted", "spoiled"
  ]);
  const [refreshing, setRefreshing] = useState(false); // for ETL refresh
  const [netTotals, setNetTotals] = useState([]);

  // Fetch species from options API
  useEffect(() => {
    fetch(BASE_API + "/api/options/species", { credentials: "include" })
      .then(async (res) => {
        if (!res.ok) throw new Error("Failed to fetch species");
        return res.json();
      })
      .then(data => {
        setSpeciesList(Array.isArray(data) ? data.filter(s => s.active).map(s => s.name) : []);
      })
      .catch(() => {
        setSpeciesList(["Chicken", "Goose", "Guinea Fowl"]);
      });
  }, []);

  // Fetch reports data
  const fetchReports = () => {
    setLoading(true);
    setError("");
    fetch(BASE_API + "/api/reports", { credentials: "include" })
      .then((r) => {
        if (!r.ok) throw new Error("Backend unavailable or DB not initialized");
        return r.json();
      })
      .then((d) => {
        setData({
          eggsOverTime: d.eggsOverTime || [],
          inventoryTrends: d.inventoryTrends || [],
          avgEggsPerCoop: d.avgEggsPerCoop || [],
          eggsByWeek: d.eggsByWeek || [],
          inventoryBySpecies: d.inventoryBySpecies || [],
          topSpecies: d.topSpecies || [],
        });
        setNetTotals(d.netTotals || []);
      })
      .catch(() => {
        setError("Backend unavailable or DB not initialized. Showing mock data.");
        setData({
          eggsOverTime: [],
          inventoryTrends: [],
          avgEggsPerCoop: [],
          eggsByWeek: [],
          inventoryBySpecies: [],
          topSpecies: [],
        });
        setNetTotals([]);
      })
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchReports();
  }, []);

  // --- ETL Refresh handler ---
  const handleFullETL = async () => {
    setRefreshing(true);
    setError("");
    try {
      const res = await fetch(BASE_API + "/api/etl/full", { method: "POST", credentials: "include" });
      if (!res.ok) throw new Error("ETL refresh failed");
      fetchReports();
    } catch (e) {
      setError("ETL refresh failed");
    } finally {
      setRefreshing(false);
    }
  };

  function SimpleBarChart({ data, xKey, yKeys, title }) {
    if (!data || data.length === 0) {
      return <div className="text-center text-gray-500 py-8">No data available.</div>;
    }
    return (
      <div className="mb-8">
        <h3 className="font-bold mb-2">{title}</h3>
        <div className="overflow-x-auto">
          <table className="min-w-full border text-sm print:text-xs">
            <thead>
              <tr className="bg-gray-100 dark:bg-gray-800">
                <th className="p-2 border">{xKey}</th>
                {yKeys.map((k) => (
                  <th key={k} className="p-2 border">{k}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {data.map((row, i) => (
                <tr key={i} className="border-b">
                  <td className="p-2 border">{row[xKey]}</td>
                  {yKeys.map((k) => (
                    <td key={k} className="p-2 border">{row[k] ?? 0}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    );
  }

  return (
    <div className="p-4 print:bg-white print:text-black">
      <h2 className="text-2xl font-bold mb-4">Reports</h2>
      <div className="flex gap-2 mb-4 print:hidden">
        <button onClick={handleFullETL} disabled={refreshing} className="px-4 py-2 bg-blue-600 text-white rounded">
          {refreshing ? "Refreshing Database..." : "Refresh"}
        </button>
        <button onClick={() => window.print()} className="px-4 py-2 bg-blue-600 text-white rounded">Print Report</button>
      </div>
      {loading ? (
        <div>Loading...</div>
      ) : (
        <>
          {error && <div className="text-yellow-600 mb-2">{error}</div>}
          {/* Net Totals Summary */}
          {netTotals.length > 0 && (
            <div className="mb-4 p-4 bg-blue-50 dark:bg-blue-900 rounded shadow flex flex-wrap gap-4">
              <div className="font-bold w-full mb-2">Net Inventory Totals by Species:</div>
              {netTotals.map(row => (
                <div key={row.species} className="px-3 py-1 bg-white dark:bg-gray-800 rounded border">
                  <span className="font-semibold">{row.species}:</span> {row.net}
                </div>
              ))}
            </div>
          )}
          <SimpleBarChart
            data={data?.eggsOverTime}
            xKey="date"
            yKeys={speciesList}
            title="Total Eggs Collected Over Time by Species"
          />
          <SimpleBarChart
            data={data?.inventoryTrends}
            xKey="date"
            yKeys={actionsList}
            title="Inventory Trends (by Action)"
          />
          <SimpleBarChart
            data={data?.avgEggsPerCoop}
            xKey="coop"
            yKeys={["avg"]}
            title="Average Eggs/Day per Coop"
          />
          <SimpleBarChart
            data={data?.eggsByWeek}
            xKey="week"
            yKeys={["count"]}
            title="Eggs Collected by Week"
          />
          <SimpleBarChart
            data={data?.inventoryBySpecies}
            xKey="species"
            yKeys={actionsList}
            title="Inventory Actions by Species"
          />
          <SimpleBarChart
            data={data?.topSpecies}
            xKey="species"
            yKeys={["total"]}
            title="Top Producing Species (Total Eggs Collected)"
          />
        </>
      )}
    </div>
  );
}

// --- Main layout with routing ---
function ProtectedRoute() {
  const { accessToken } = useAuth();
  if (!accessToken) {
    return <Navigate to="/login" replace />;
  }
  return <Outlet />;
}

function MainLayout() {
  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />
      <main className="flex-1">
        <Routes>
          <Route element={<ProtectedRoute />}>
            <Route path="/inventory" element={<InventoryPage />} />
            <Route path="/options" element={<OptionsPage />} />
            <Route path="/reports" element={<ReportsPage />} />
          </Route>
          <Route path="*" element={<Navigate to="/inventory" replace />} />
        </Routes>
      </main>
    </div>
  );
}

// --- Combined Auth Page ---
function AuthPage() {
  return (
    <div className="flex flex-col md:flex-row gap-8 items-start justify-center min-h-screen p-8 bg-gray-50 dark:bg-gray-900">
      <div className="flex-1 flex justify-center"><LoginForm /></div>
      <div className="flex-1 flex justify-center"><RegisterForm /></div>
    </div>
  );
}

// --- Login/Register Forms ---
function LoginForm() {
  const { login, loading, error } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [formError, setFormError] = useState(null);
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setFormError(null);
    if (!email || !password) {
      setFormError("Email and password required");
      return;
    }
    await login(email, password);
    if (!error) {
      navigate("/inventory");
    }
  };

  return (
    <form className="flex flex-col gap-4 w-80 bg-white dark:bg-gray-800 p-6 rounded shadow" onSubmit={handleSubmit}>
      <h2 className="text-xl font-bold">Login</h2>
      <input
        type="email"
        placeholder="Email"
        className="p-2 border rounded"
        value={email}
        onChange={e => setEmail(e.target.value)}
        required
      />
      <input
        type="password"
        placeholder="Password"
        className="p-2 border rounded"
        value={password}
        onChange={e => setPassword(e.target.value)}
        required
      />
      {(formError || error) && <div className="text-red-500">{formError || error}</div>}
      <button type="submit" className="bg-blue-600 text-white py-2 rounded" disabled={loading}>
        {loading ? "Logging in..." : "Login"}
      </button>
    </form>
  );
}

function RegisterForm() {
  const { register, loading, error } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [formError, setFormError] = useState(null);
  const navigate = useNavigate();

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
    if (!error) {
      navigate("/inventory");
    }
  };

  return (
    <form className="flex flex-col gap-4 w-80 bg-white dark:bg-gray-800 p-6 rounded shadow" onSubmit={handleSubmit}>
      <h2 className="text-xl font-bold">Register</h2>
      <input
        type="email"
        placeholder="Email"
        className="p-2 border rounded"
        value={email}
        onChange={e => setEmail(e.target.value)}
        required
      />
      <input
        type="password"
        placeholder="Password (min 8 chars)"
        className="p-2 border rounded"
        value={password}
        onChange={e => setPassword(e.target.value)}
        required
      />
      {(formError || error) && <div className="text-red-500">{formError || error}</div>}
      <button type="submit" className="bg-green-600 text-white py-2 rounded" disabled={loading}>
        {loading ? "Registering..." : "Register"}
      </button>
    </form>
  );
}

// --- Simple router for demo ---
function App() {
  const { accessToken } = useAuth();

  return (
    <Router>
      <Routes>
        <Route path="/login" element={<AuthPage />} />
        {/* Remove /register route, handled by AuthPage */}
        <Route element={<ProtectedRoute />}>
          <Route path="/*" element={<MainLayout />} />
        </Route>
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    </Router>
  );
}

ReactDOM.createRoot(document.getElementById("root")).render(
  <ThemeProvider>
    <AuthProvider>
      <App />
    </AuthProvider>
  </ThemeProvider>
);