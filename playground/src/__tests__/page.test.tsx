import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import PlaygroundPage from "@/app/page";
import { DiffResult } from "@/lib/types";

vi.mock("@monaco-editor/react", () => ({
  default: ({ value, onChange }: { value: string; onChange: (v: string) => void }) => (
    <textarea
      data-testid="monaco-editor"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  ),
}));

const mockDiffResult: DiffResult = {
  base_file: "",
  head_file: "",
  summary: { total: 2, breaking: 1, non_breaking: 1, info: 0 },
  changes: [
    { type: "endpoint_removed", severity: "breaking", path: "/users", method: "DELETE", location: "", description: "Endpoint removed" },
    { type: "field_added", severity: "non-breaking", path: "/users", method: "POST", location: "", description: "Field added" },
  ],
};

beforeEach(() => {
  vi.resetAllMocks();
});

describe("PlaygroundPage", () => {
  it("renders all three schema type tabs", () => {
    render(<PlaygroundPage />);
    expect(screen.getByRole("button", { name: "OpenAPI" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "GraphQL" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "gRPC / Protobuf" })).toBeInTheDocument();
  });

  it("has OpenAPI active by default", () => {
    render(<PlaygroundPage />);
    const openApiBtn = screen.getByRole("button", { name: "OpenAPI" });
    expect(openApiBtn.className).toMatch(/bg-indigo-600/);
  });

  it("switches active tab when clicked", () => {
    render(<PlaygroundPage />);
    const graphqlBtn = screen.getByRole("button", { name: "GraphQL" });
    fireEvent.click(graphqlBtn);
    expect(graphqlBtn.className).toMatch(/bg-indigo-600/);
    expect(screen.getByRole("button", { name: "OpenAPI" }).className).not.toMatch(/bg-indigo-600/);
  });

  it("renders the Compare button", () => {
    render(<PlaygroundPage />);
    expect(screen.getByRole("button", { name: "Compare" })).toBeInTheDocument();
  });

  it("renders the Generate spec link pointing to docs", () => {
    render(<PlaygroundPage />);
    const link = screen.getByRole("link", { name: /generate spec/i });
    expect(link).toBeInTheDocument();
    expect(link).toHaveAttribute("href", "https://driftabot.github.io/driftabot-engine/generating-specs");
    expect(link).toHaveAttribute("target", "_blank");
  });

  it("calls fetch with correct payload on Compare click", async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockDiffResult,
    });
    vi.stubGlobal("fetch", mockFetch);

    render(<PlaygroundPage />);
    fireEvent.click(screen.getByRole("button", { name: "Compare" }));

    await waitFor(() => expect(mockFetch).toHaveBeenCalledOnce());

    const [url, options] = mockFetch.mock.calls[0];
    expect(url).toMatch(/\/api\/compare/);
    const body = JSON.parse(options.body);
    expect(body.schema_type).toBe("openapi");
    expect(body.base_content).toBeTruthy();
    expect(body.head_content).toBeTruthy();
  });

  it("shows error banner when API returns an error", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: false,
      json: async () => ({ error: "invalid schema" }),
    }));

    render(<PlaygroundPage />);
    fireEvent.click(screen.getByRole("button", { name: "Compare" }));

    await waitFor(() => screen.getByText("invalid schema"));
    expect(screen.getByText("invalid schema")).toBeInTheDocument();
  });

  it("shows results table on successful compare", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockDiffResult,
    }));

    render(<PlaygroundPage />);
    fireEvent.click(screen.getByRole("button", { name: "Compare" }));

    await waitFor(() => screen.getByText("2 total"));
    expect(screen.getByText("2 total")).toBeInTheDocument();
    expect(screen.getByText("1 breaking")).toBeInTheDocument();
    expect(screen.getByText("BREAKING")).toBeInTheDocument();
  });

  it("shows error banner on network failure", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("Network error")));

    render(<PlaygroundPage />);
    fireEvent.click(screen.getByRole("button", { name: "Compare" }));

    await waitFor(() => screen.getByText("Network error"));
    expect(screen.getByText("Network error")).toBeInTheDocument();
  });
});
