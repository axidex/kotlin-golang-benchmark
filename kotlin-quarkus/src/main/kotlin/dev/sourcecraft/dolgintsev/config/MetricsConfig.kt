package dev.sourcecraft.dolgintsev.config

import io.micrometer.core.instrument.Meter
import io.micrometer.core.instrument.config.MeterFilter
import io.micrometer.core.instrument.distribution.DistributionStatisticConfig
import jakarta.enterprise.inject.Produces
import jakarta.inject.Singleton

class MetricsConfig {

    @Produces
    @Singleton
    fun enableHistogram(): MeterFilter {
        return object : MeterFilter {
            override fun configure(id: Meter.Id, config: DistributionStatisticConfig): DistributionStatisticConfig {
                if (id.name.startsWith("http.server.requests")) {
                    return DistributionStatisticConfig.builder()
                        .percentiles(0.5, 0.95, 0.99)
                        .percentilesHistogram(true)
                        .build()
                        .merge(config)
                }
                return config
            }
        }
    }
}