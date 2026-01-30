-- Seed templates for Zapiki
-- Run this after creating circuits

-- First, we need to create the circuits that templates will use
-- These circuits should already exist from Phase 3

-- Template 1: Age Verification (18+)
INSERT INTO templates (
    id,
    name,
    description,
    category,
    proof_system,
    circuit_id,
    input_schema,
    example_inputs,
    documentation,
    is_active
) VALUES (
    uuid_generate_v4(),
    'Age Verification (18+)',
    'Prove that you are 18 years or older without revealing your actual age',
    'Identity',
    'groth16',
    (SELECT id FROM circuits WHERE name = 'Age Verification Circuit' LIMIT 1),
    '{"type":"object","required":["age","min_age","is_adult"],"properties":{"age":{"type":"number","description":"Your actual age (kept secret)","minimum":0,"maximum":150},"min_age":{"type":"number","description":"Minimum age requirement","default":18},"is_adult":{"type":"number","description":"Set to 1 to prove you meet the requirement","enum":[0,1]}}}',
    '{"age":25,"min_age":18,"is_adult":1}',
    '# Age Verification Template

This template allows you to prove that you are above a certain age (default: 18) without revealing your actual age.

## Use Cases
- Access age-restricted content
- Verify eligibility for services
- Anonymous age verification

## Inputs
- **age** (secret): Your actual age
- **min_age** (public): The minimum age requirement (usually 18)
- **is_adult** (public): Set to 1 to indicate you meet the requirement

## Privacy
Your actual age remains secret. The proof only reveals that you meet the minimum age requirement.

## Example
```json
{
  "age": 25,
  "min_age": 18,
  "is_adult": 1
}
```

This proves you are >= 18, but doesn''t reveal you are 25.',
    true
) ON CONFLICT DO NOTHING;

-- Template 2: Age Verification (21+)
INSERT INTO templates (
    id,
    name,
    description,
    category,
    proof_system,
    circuit_id,
    input_schema,
    example_inputs,
    documentation,
    is_active
) VALUES (
    uuid_generate_v4(),
    'Age Verification (21+)',
    'Prove that you are 21 years or older without revealing your actual age',
    'Identity',
    'groth16',
    (SELECT id FROM circuits WHERE name = 'Age Verification Circuit' LIMIT 1),
    '{"type":"object","required":["age","min_age","is_adult"],"properties":{"age":{"type":"number","description":"Your actual age (kept secret)","minimum":0,"maximum":150},"min_age":{"type":"number","description":"Minimum age requirement","default":21},"is_adult":{"type":"number","description":"Set to 1 to prove you meet the requirement","enum":[0,1]}}}',
    '{"age":25,"min_age":21,"is_adult":1}',
    '# Age Verification Template (21+)

Prove you are 21 or older for services that require this age threshold.

## Example
```json
{
  "age": 28,
  "min_age": 21,
  "is_adult": 1
}
```',
    true
) ON CONFLICT DO NOTHING;

-- Template 3: Salary Range Proof
INSERT INTO templates (
    id,
    name,
    description,
    category,
    proof_system,
    circuit_id,
    input_schema,
    example_inputs,
    documentation,
    is_active
) VALUES (
    uuid_generate_v4(),
    'Salary Range Verification',
    'Prove your salary falls within a specific range without revealing the exact amount',
    'Financial',
    'groth16',
    (SELECT id FROM circuits WHERE name = 'Range Proof Circuit' LIMIT 1),
    '{"type":"object","required":["value","min","max","in_range"],"properties":{"value":{"type":"number","description":"Your actual salary (kept secret)"},"min":{"type":"number","description":"Minimum salary range"},"max":{"type":"number","description":"Maximum salary range"},"in_range":{"type":"number","description":"Set to 1 to prove salary is in range","enum":[0,1]}}}',
    '{"value":75000,"min":50000,"max":100000,"in_range":1}',
    '# Salary Range Verification

Prove your salary is within a certain range without revealing the exact amount.

## Use Cases
- Loan applications
- Financial verification
- Anonymous salary surveys

## Inputs
- **value** (secret): Your actual salary
- **min** (public): Minimum of acceptable range
- **max** (public): Maximum of acceptable range
- **in_range** (public): Set to 1 to prove value is in range

## Privacy
Your exact salary remains secret. Only proves it falls within [min, max].

## Example
```json
{
  "value": 75000,
  "min": 50000,
  "max": 100000,
  "in_range": 1
}
```

This proves your salary is between $50k-$100k without revealing it''s $75k.',
    true
) ON CONFLICT DO NOTHING;

-- Template 4: Credit Score Range
INSERT INTO templates (
    id,
    name,
    description,
    category,
    proof_system,
    circuit_id,
    input_schema,
    example_inputs,
    documentation,
    is_active
) VALUES (
    uuid_generate_v4(),
    'Credit Score Range Verification',
    'Prove your credit score is within a certain range without revealing the exact score',
    'Financial',
    'groth16',
    (SELECT id FROM circuits WHERE name = 'Range Proof Circuit' LIMIT 1),
    '{"type":"object","required":["value","min","max","in_range"],"properties":{"value":{"type":"number","description":"Your credit score (kept secret)","minimum":300,"maximum":850},"min":{"type":"number","description":"Minimum acceptable score"},"max":{"type":"number","description":"Maximum score range"},"in_range":{"type":"number","description":"Set to 1","enum":[1]}}}',
    '{"value":720,"min":700,"max":850,"in_range":1}',
    '# Credit Score Range Verification

Prove your credit score meets requirements without revealing the exact score.

## Example
```json
{
  "value": 720,
  "min": 700,
  "max": 850,
  "in_range": 1
}
```

Proves score is >= 700 without revealing it''s 720.',
    true
) ON CONFLICT DO NOTHING;

-- Template 5: Simple Multiplication Proof
INSERT INTO templates (
    id,
    name,
    description,
    category,
    proof_system,
    circuit_id,
    input_schema,
    example_inputs,
    documentation,
    is_active
) VALUES (
    uuid_generate_v4(),
    'Multiplication Proof',
    'Prove you know two numbers that multiply to a specific result',
    'Math',
    'groth16',
    (SELECT id FROM circuits WHERE name = 'Simple Circuit' LIMIT 1),
    '{"type":"object","required":["x","y","z"],"properties":{"x":{"type":"number","description":"First number (kept secret)"},"y":{"type":"number","description":"Second number (kept secret)"},"z":{"type":"number","description":"Result (public)"}}}',
    '{"x":7,"y":8,"z":56}',
    '# Multiplication Proof

Prove you know two numbers that multiply to a given result without revealing the factors.

## Use Cases
- Cryptographic challenges
- Math puzzles
- Factor knowledge proofs

## Example
```json
{
  "x": 7,
  "y": 8,
  "z": 56
}
```

Proves you know factors of 56 without revealing they are 7 and 8.',
    true
) ON CONFLICT DO NOTHING;

-- Display inserted templates
SELECT
    name,
    category,
    proof_system,
    is_active
FROM templates
ORDER BY category, name;
