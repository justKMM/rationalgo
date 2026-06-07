export const BUDGET_TIERS = {
  cheapass: { label: "Cheapass", eur: 5 },
  mid: { label: "Mid", eur: 10 },
  luxury: { label: "Luxury Pro VIP", eur: 15 },
} as const;

export type BudgetTier = keyof typeof BUDGET_TIERS;

export function budgetEur(tier: BudgetTier): number {
  return BUDGET_TIERS[tier].eur;
}
