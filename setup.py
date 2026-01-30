"""Setup script for pulumi-lagoon-provider."""

from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setup(
    name="pulumi-lagoon",
    version="0.1.0",
    author="Greg Chaix",
    author_email="greg@tag1consulting.com",
    description="A Pulumi dynamic provider for managing Lagoon resources",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/tag1consulting/pulumi-lagoon-provider",
    packages=find_packages(exclude=["tests", "examples"]),
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "Intended Audience :: System Administrators",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: System :: Systems Administration",
        "License :: OSI Approved :: Apache Software License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
    ],
    python_requires=">=3.8",
    install_requires=[
        "pulumi>=3.0.0,<4.0.0",
        "requests>=2.28.0,<3.0.0",
        "PyJWT>=2.8.0",  # For admin JWT token generation
    ],
    extras_require={
        "dev": [
            "pytest>=7.0.0",
            "pytest-cov>=4.0.0",
            "pytest-mock>=3.10.0",
            "ruff>=0.1.0",
            "mypy>=1.0.0",
            "types-requests>=2.28.0",
        ],
    },
    keywords="pulumi lagoon infrastructure-as-code iac devops kubernetes",
    project_urls={
        "Bug Reports": "https://github.com/tag1consulting/pulumi-lagoon-provider/issues",
        "Source": "https://github.com/tag1consulting/pulumi-lagoon-provider",
        "Documentation": "https://github.com/tag1consulting/pulumi-lagoon-provider/blob/main/README.md",
    },
)
