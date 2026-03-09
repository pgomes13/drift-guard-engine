import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import ResultsTable from "@/components/ResultsTable";
import { DiffResult } from "@/lib/types";

const baseResult: DiffResult = {
  base_file: "",
  head_file: "",
  summary: { total: 0, breaking: 0, non_breaking: 0, info: 0 },
  changes: [],
};

describe("ResultsTable", () => {
  it("shows 'No changes detected' when changes is empty", () => {
    render(<ResultsTable result={baseResult} />);
    expect(screen.getByText(/No changes detected/i)).toBeInTheDocument();
  });

  it("renders total pill", () => {
    const result: DiffResult = {
      ...baseResult,
      summary: { total: 3, breaking: 1, non_breaking: 1, info: 1 },
      changes: [
        { type: "endpoint_removed", severity: "breaking", path: "/users", method: "DELETE", location: "", description: "Endpoint removed" },
        { type: "field_added", severity: "non-breaking", path: "/users", method: "POST", location: "", description: "Field added" },
        { type: "description_changed", severity: "info", path: "/users", method: "GET", location: "", description: "Description changed" },
      ],
    };
    render(<ResultsTable result={result} />);
    expect(screen.getByText("3 total")).toBeInTheDocument();
  });

  it("renders breaking, non-breaking, and info pills when counts > 0", () => {
    const result: DiffResult = {
      ...baseResult,
      summary: { total: 3, breaking: 1, non_breaking: 1, info: 1 },
      changes: [
        { type: "endpoint_removed", severity: "breaking", path: "/a", method: "DELETE", location: "", description: "Endpoint removed" },
        { type: "field_added", severity: "non-breaking", path: "/b", method: "POST", location: "", description: "Field added" },
        { type: "description_changed", severity: "info", path: "/c", method: "GET", location: "", description: "Description changed" },
      ],
    };
    render(<ResultsTable result={result} />);
    expect(screen.getByText("1 breaking")).toBeInTheDocument();
    expect(screen.getByText("1 non-breaking")).toBeInTheDocument();
    expect(screen.getByText("1 info")).toBeInTheDocument();
  });

  it("hides pills with zero counts", () => {
    const result: DiffResult = {
      ...baseResult,
      summary: { total: 1, breaking: 1, non_breaking: 0, info: 0 },
      changes: [
        { type: "endpoint_removed", severity: "breaking", path: "/a", method: "DELETE", location: "", description: "Endpoint removed" },
      ],
    };
    render(<ResultsTable result={result} />);
    expect(screen.queryByText(/non-breaking/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/\d+ info/)).not.toBeInTheDocument();
  });

  it("renders severity badges in the table", () => {
    const result: DiffResult = {
      ...baseResult,
      summary: { total: 2, breaking: 1, non_breaking: 1, info: 0 },
      changes: [
        { type: "endpoint_removed", severity: "breaking", path: "/a", method: "DELETE", location: "", description: "Endpoint removed" },
        { type: "field_added", severity: "non-breaking", path: "/b", method: "POST", location: "", description: "Field added" },
      ],
    };
    render(<ResultsTable result={result} />);
    expect(screen.getByText("BREAKING")).toBeInTheDocument();
    expect(screen.getByText("NON-BREAKING")).toBeInTheDocument();
  });

  it("renders before/after values when present", () => {
    const result: DiffResult = {
      ...baseResult,
      summary: { total: 1, breaking: 1, non_breaking: 0, info: 0 },
      changes: [
        {
          type: "type_changed",
          severity: "breaking",
          path: "/users/{id}",
          method: "GET",
          location: "",
          description: "Type changed",
          before: "string",
          after: "integer",
        },
      ],
    };
    render(<ResultsTable result={result} />);
    expect(screen.getByText("string")).toBeInTheDocument();
    expect(screen.getByText(/integer/)).toBeInTheDocument();
  });

  it("renders table headers", () => {
    const result: DiffResult = {
      ...baseResult,
      summary: { total: 1, breaking: 1, non_breaking: 0, info: 0 },
      changes: [
        { type: "endpoint_removed", severity: "breaking", path: "/a", method: "DELETE", location: "", description: "Removed" },
      ],
    };
    render(<ResultsTable result={result} />);
    expect(screen.getByText("Severity")).toBeInTheDocument();
    expect(screen.getByText("Type")).toBeInTheDocument();
    expect(screen.getByText("Path")).toBeInTheDocument();
    expect(screen.getByText("Description")).toBeInTheDocument();
  });
});
