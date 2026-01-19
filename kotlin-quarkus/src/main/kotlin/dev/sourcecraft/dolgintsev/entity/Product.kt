package dev.sourcecraft.dolgintsev.entity

import io.quarkus.hibernate.orm.panache.kotlin.PanacheCompanion
import io.quarkus.hibernate.orm.panache.kotlin.PanacheEntity
import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.Table
import java.math.BigDecimal

@Entity
@Table(name = "products")
class Product : PanacheEntity() {
    companion object : PanacheCompanion<Product>

    @Column(nullable = false)
    lateinit var name: String

    @Column(length = 1000)
    var description: String? = null

    @Column(nullable = false)
    var price: BigDecimal = BigDecimal.ZERO

    @Column(nullable = false)
    var quantity: Int = 0
}
