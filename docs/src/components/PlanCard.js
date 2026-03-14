import React from 'react';
import PropTypes from 'prop-types';

function PlanCard({ name, price, features, cta }) {
  return (
    <div className="plan-card">
      <div className="plan-card-header">
        <span className="plan-card-name">{name}</span>
        <span className="plan-card-price">{price}</span>
      </div>
      <ul className="plan-card-features">
        {features.map((feature, index) => (
          <li key={index} className="plan-card-feature">
            <span className="plan-card-feature-label">{feature.label}</span>
            <span className="plan-card-feature-value">{feature.value}</span>
          </li>
        ))}
      </ul>
      {cta && (
        <a className="plan-card-cta" href={cta.link}>
          {cta.text}
        </a>
      )}
    </div>
  );
}

PlanCard.propTypes = {
  name: PropTypes.string.isRequired,
  price: PropTypes.string.isRequired,
  features: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.string.isRequired,
      value: PropTypes.string.isRequired,
    })
  ).isRequired,
  cta: PropTypes.shape({
    text: PropTypes.string.isRequired,
    link: PropTypes.string.isRequired,
  }),
};

function PlanCards({ children }) {
  return <div className="plan-cards">{children}</div>;
}

export { PlanCard, PlanCards };
export default PlanCard;
